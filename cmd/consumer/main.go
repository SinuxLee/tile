package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/sinuxlee/tile/internal/entity"
	"github.com/sinuxlee/tile/pkg/disruptor"
)

const (
	serverName   = "consumer"
	nodeID       = 1 // 保证同类型的进程编号不相同
	msgQueueName = "push_stream"
)

func main() {
	//cluster := redis.NewClusterClient(&redis.ClusterOptions{
	//	Addrs:        []string{"127.0.0.1:7000"},
	//	ReadTimeout:  time.Second * 60,
	//	WriteTimeout: time.Second * 60,
	//})

	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6379",
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
	})

	cn, err := disruptor.NewConsumer(&disruptor.ConsumerOptions{
		Consumer:  fmt.Sprintf("%v_%v", serverName, nodeID),
		QueueName: msgQueueName,
	}, client)

	if err != nil {
		panic(err)
	}

	p := &entity.Player{}
	more := true
	for more {
		_, more = cn.Pop(p, func(_ disruptor.Message) error {
			println(p.UserId, p.NickName)
			return nil
		})
	}

	// 监听信号
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch

	cn.Close()

}
