package common

import (
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func SetupCloseHandler(c *Cleaner) (ch chan os.Signal) {
	ch = make(chan os.Signal)
	// see https://en.wikipedia.org/wiki/Signal_(IPC)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		<-ch
		if err := (*c).Clean(); err != nil {
			log.Err(err)
		}
		os.Exit(0)
	}()
	return
}
