package sub

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

//https://www.redis.net.cn/order/
//https://blog.csdn.net/qq_31960623/article/details/117911710
var ctx = context.Background()

func client() (client *redis.Client) {
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "pwd", // no password set
		DB:       0,     // use default DB
		PoolSize: 10,    //默认一个CPU 10个连接
	})
	return client
}

// string type
func TestString(t *testing.T) {
	client := client()
	defer client.Close()
	_, err := client.Set(ctx, "name", "wzj", 10*time.Second).Result()
	handle(err, "failed to set string for key 'name'")

	val, err := client.Get(ctx, "name").Result()
	handle(err, "failed to get value for key 'name'")
	assert.Equal(t, "wzj", val, "value doesn't equal with val")
}

func TestStringTimeOut(t *testing.T) {
	client := client()
	defer client.Close()

	_, err := client.Set(ctx, "name", "wzj", 2*time.Second).Result()
	handle(err, "failed to set string for key 'name'")

	time.Sleep(3 * time.Second)

	//超过有效期，取得的值为空
	_, err = client.Get(ctx, "name").Result()
	assert.Error(t, err, "the key 'name' shouldn't be existed")
}

//key:string, value: int/byte/bool/float32/...
/**
String是最简单也是最常用的数据类型，通过set和get方法设置或获取数据，有如下使用场景

**缓存功能：**最常用的功能，没有之一。比如，对某个用户对象转成JSON字符串，读取后再转换回目标对象；
**计数器：**常用于限制某个接口的请求次数，或者统计用户的点击次数等等，使用incr命令实现自增。实现计数器
*/
func TestOtherValues(t *testing.T) {
	client := client()
	defer client.Close()

	//Redis SET 命令用于设置给定 key 的值。如果 key 已经存储其他值， SET 就覆写旧值，且无视类型。
	//int
	_, err := client.Set(ctx, "intVal", 23, 50*time.Second).Result()
	handle(err, "failed to set int for key 'intVal'")
	val, err := client.Get(ctx, "intVal").Int()
	assert.Equal(t, val, 23, "intVal should be 32")

	//float值
	_, err = client.Set(ctx, "float32Value", 43.67, 50*time.Second).Result()
	handle(err, "failed to set float32 for key 'float32Value'")
	valFloat32, err := client.Get(ctx, "float32Value").Float32()
	println(valFloat32)

	//bool值
	client.Set(ctx, "boolVal", true, 0)
	boolVal, _ := client.Get(ctx, "boolVal").Bool()
	assert.Equal(t, boolVal, true, "the value should be true")

	//增加
	initialVal := 1
	client.Set(ctx, "initialVal", initialVal, 0)
	client.Incr(ctx, "initialVal")
	client.IncrBy(ctx, "initialVal", 3)

	//减值
	client.Decr(ctx, "initialVal")
	client.DecrBy(ctx, "initialVal", 2)

	//iVal, err := client.Get(ctx, "initialVal").Int()
	//assert.Equal(t, iVal, 5, "the initialVal should be 5")

	//一开始不存在的key，进行增加
	client.Incr(ctx, "noneExist")
	client.Incr(ctx, "noneExist")
	client.Incr(ctx, "noneExist")
	client.Expire(ctx, "noneExist", 30*time.Second)

	//返回多个key
	array, err := client.MGet(ctx, "initialVal", "boolVal", "none").Result()
	for _, val := range array {
		if val != nil {
			println(fmt.Sprintf("val=%s", val))
		}
	}

	//set only if the key doesn't exist, similar with lock
	result, err := client.SetNX(ctx, "lockBy", "me", 60*time.Second).Result()
	handle(err, "SetNX failed")
	println(result)

	//设置多个key
	client.MSet(ctx, "key1", "value1", "key2", "value2")

	//当不存在时，set多个key
	client.MSetNX(ctx, "lock1", "lock1", "lock2", "lock2")

	//为指定的 key 设置值及其过期时间。如果 key 已经存在， SETEX 命令将会替换旧的值
	client.SetEx(ctx, "exKey", "valueEx", 60*time.Second)

	//获取指定 key 所储存的字符串值的长度。当 key 储存的不是字符串值时，返回一个错误。
	i, err := client.StrLen(ctx, "exKey").Result()
	println("exKey.len=", i)

	//如果 key 已经存在并且是一个字符串， APPEND 命令将 value 追加到 key 原来的值的末尾。
	client.Append(ctx, "exKey", ":AppendValue")
	s, err := client.Get(ctx, "exKey").Result()
	println("exKey=", s)

	//命令用于设置指定 key 的值，并返回 key 旧的值。
	empty, err := client.GetSet(ctx, "oldKey", "newValue").Result()
	s3, err := client.GetSet(ctx, "oldKey", "newValue").Result()
	println(empty, "==", s3)

}

func handle(err error, msg string) {
	if err != nil {
		println(msg)
	}
}
