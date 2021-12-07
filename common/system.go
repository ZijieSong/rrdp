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
			log.Error().Msgf(err.Error())
		}
		os.Exit(0)
	}()
	return
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
