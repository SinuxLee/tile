package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/sinuxlee/tile/internal/entity"
	"github.com/sinuxlee/tile/pkg/disruptor"
)

const (
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
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
	})

	pr, err := disruptor.NewProducer(&disruptor.ProducerOptions{
		QueueName: msgQueueName,
	}, client)

	if err != nil {
		panic(err)
	}

	p := &entity.Player{
		AccessToken:   "5876dd81-bb6e-5972-9b5a-832333d5b0ee",
		RefreshToken:  "4fe2c2d0-bfb7-500d-9cf6-f5c4f4bcaabb",
		HeadImgUrl:    "https://www.baidu.com/img/PCtm_d9c8750bed0b3c7d089fa7d55720d6cf.png",
		UserId:        131198216,
		ExpiresIn:     14400,
		LastLoginTime: time.Now().Unix(),
	}

	inx := int64(0)
	r := gin.Default()
	r.GET("/", func(ctx *gin.Context) {
		p.NickName = fmt.Sprintf("robot_%v", atomic.AddInt64(&inx, 1))
		p.UserIp = ctx.Request.RemoteAddr
		pr.Push(p)
	})

	r.Run(":8090")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch

	pr.Close()
}
