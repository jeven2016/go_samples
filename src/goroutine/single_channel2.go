package goroutine

import (
	"runtime"
	"strconv"
)

//创建一个容量为3的chanel
var chanel3 = make(chan string, 3)

//限制为单项通道，表示通道只能接受消息
func send3(ch chan<- string) {
	for i := 0; i < 3; i++ {
		chanel <- strconv.Itoa(i)
	}
	println("msg sent")
	//发送完后就关闭channel， 其中的数据仍然可以被消费到
	close(ch)
	println("channel closed")
}

//限制为单项通道，表示只能从通道获取数据，不能写入
func receive3(ch <-chan string) {
	println("receiver started...")
	//1) 如果写端没有写数据，也没有关闭。<-ch; 会阻塞
	//2）如果写端写数据， value 保存 <-ch 读到的数据。 ok 被设置为 true
	//3）如果写端关闭。 value 为数据类型默认值。ok 被设置为 false
	for {
		//channel没有关闭的情况下，如果没有数据会阻塞
		if msg, ok := <-chanel2; ok {
			println("receiver consumes a msg :", msg)
		} else {
			//chanel关闭则中断
			break
		}
	}
}

func RunSingleChannel() {
	go send3(chanel3)
	go receive3(chanel3)

	runtime.Gosched()
}
