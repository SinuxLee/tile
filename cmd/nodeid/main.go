package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/sinuxlee/tile/internal/app"
	"github.com/sinuxlee/tile/pkg/logger"
)

// 常量定义
const (
	serviceID   = 110
	serviceName = "nodeid"
)

func main() {
	srv, err := app.New(
		app.ServiceID(serviceID),
		app.ServiceName(serviceName),
		app.Conf(),
		app.LogLevel(),
		app.Named(),
		app.Dao(),
		app.UseCase(),
		app.Controller(),
		app.Router(),
		app.PProf(),
		app.HTTPServer(),
	)

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	log.Info().Msg("app init success")

	// 监听信号
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	_ = srv.Run(ch)

	<-ch

	_ = logger.Close()
}
