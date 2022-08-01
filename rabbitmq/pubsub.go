package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"flag"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"log"
	"os"
)

// Command pubsub is an example of a fanout exchange with dynamic reliable
// membership, reading from stdin, writing to stdout.
//
// This example shows how to implement reconnect logic independent from a
// publish/subscribe loop with bridges to application types.

//var url = flag.String("url", "amqp:///", "AMQP url for both the publisher and subscriber")
var url string = "amqp://admin:admin@localhost:5672"

// exchange binds the publishers to the subscribers
const exchange = "pubsub"

// message is the application type for a message.  This can contain identity,
// or a reference to the receiver chan for further demuxing.
type message []byte

// session composes an amqp.Connection with an amqp.Channel
type session struct {
	*amqp.Connection
	*amqp.Channel
}

// Close tears the connection down, taking the channel with it.
func (s session) Close() error {
	if s.Connection == nil {
		return nil
	}
	return s.Connection.Close()
}

// redial continually connects to the URL, exiting the program when no longer possible
func redial(ctx context.Context, url string) chan chan session {
	sessions := make(chan chan session)

	go func() {
		sess := make(chan session)
		defer close(sessions)

		for {
			select {
			case sessions <- sess:
			case <-ctx.Done():
				log.Println("shutting down session factory")
				return
			}

			conn, err := amqp.Dial(url)
			if err != nil {
				log.Printf("cannot (re)dial: %v: %q", err, url)
				continue
			}

			ch, err := conn.Channel()
			if err != nil {
				log.Printf("cannot create channel: %v", err)
				continue
			}

			if err := ch.ExchangeDeclare(exchange, "fanout", false, true, false, false, nil); err != nil {
				log.Printf("cannot declare fanout exchange: %v", err)
				continue
			}

			select {
			case sess <- session{conn, ch}:
			case <-ctx.Done():
				log.Println("shutting down new session")
				return
			}
		}
	}()

	return sessions
}

// publish publishes messages to a reconnecting session to a fanout exchange.
// It receives from the application specific source of messages.
func publish(sessions chan chan session, messages <-chan message) {
	for session := range sessions {
		var (
			running bool
			reading = messages
			pending = make(chan message, 1)
			confirm = make(chan amqp.Confirmation, 1)
		)

		pub := <-session

		// publisher confirms for this channel/connection
		if err := pub.Confirm(false); err != nil {
			log.Printf("publisher confirms not supported")
			close(confirm) // confirms not supported, simulate by always nacking
		} else {
			pub.NotifyPublish(confirm)
		}

		log.Printf("publishing...")

	Publish:
		for {
			var body message
			select {
			case confirmed, ok := <-confirm:
				if !ok {
					break Publish
				}
				if !confirmed.Ack {
					log.Printf("nack message %d, body: %q", confirmed.DeliveryTag, string(body))
				}
				reading = messages

			case body = <-pending:
				routingKey := "ignored for fanout exchanges, application dependent for other exchanges"
				err := pub.Publish(exchange, routingKey, false, false, amqp.Publishing{
					Body: body,
				})
				// Retry failed delivery on the next session
				if err != nil {
					pending <- body
					pub.Close()
					break Publish
				}

			case body, running = <-reading:
				// all messages consumed
				if !running {
					return
				}
				// work on pending delivery until ack'd
				pending <- body
				reading = nil
			}
		}
	}
}

// identity returns the same host/process unique string for the lifetime of
// this process so that subscriber reconnections reuse the same queue name.
func identity() string {
	hostname, err := os.Hostname()
	h := sha1.New()
	fmt.Fprint(h, hostname)
	fmt.Fprint(h, err)
	fmt.Fprint(h, os.Getpid())
	//return fmt.Sprintf("%x", h.Sum(nil))
	return "hello-queue"
}

// subscribe consumes deliveries from an exclusive queue from a fanout exchange and sends to the application specific messages chan.
func subscribe(sessions chan chan session, messages chan<- message) {
	queue := "hello-queue"

	for session := range sessions {
		sub := <-session

		if _, err := sub.QueueDeclare(queue, false, false, false, false, nil); err != nil {
			log.Printf("cannot consume from exclusive queue: %q, %v", queue, err)
			continue
		}

		routingKey := "hello-routing-key"
		if err := sub.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
			log.Printf("cannot consume without a binding to exchange: %q, %v", exchange, err)
			continue
		}

		deliveries, err := sub.Consume(queue, "", false, false, false, false, nil)
		if err != nil {
			log.Printf("cannot consume from: %q, %v", queue, err)
			continue
		}

		log.Printf("subscribed...")

		for msg := range deliveries {
			log.Printf("msg received: %v", msg.Body)
			messages <- msg.Body
			sub.Ack(msg.DeliveryTag, false)
		}
	}
}

// read is this application's translation to the message format, scanning from
// stdin.
func read(r io.Reader) <-chan message {
	lines := make(chan message)
	go func() {
		defer close(lines)
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			lines <- scan.Bytes()
		}
	}()
	return lines
}

// write is this application's subscriber of application messages, printing to
// stdout.
func write(w io.Writer) chan<- message {
	lines := make(chan message)
	go func() {
		for line := range lines {
			fmt.Fprintln(w, string(line))
		}
	}()
	return lines
}

func main() {
	flag.Parse()

	ctx, done := context.WithCancel(context.Background())

	go func() {
		publish(redial(ctx, url), read(os.Stdin))
		done()
	}()

	go func() {
		subscribe(redial(ctx, url), write(os.Stdout))
		done()
	}()

	<-ctx.Done()
}