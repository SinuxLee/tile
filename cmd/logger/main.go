package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sinuxlee/tile/pkg/logger"
)

func main() {
	logger.Init(logger.File(true))
	for i := 0; i < 10; i++ {
		go func(i int) {
			for {
				log.Info().Int("id", i).Msg("test")
				time.Sleep(time.Second)
			}

		}(i)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch
	logger.Close()
	fmt.Fprintf(os.Stderr, "game over")
}
