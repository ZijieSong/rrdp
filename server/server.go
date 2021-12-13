package main

import (
	"github.com/ZijieSong/rrdp/common"
	"github.com/ZijieSong/rrdp/pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"sync"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	listen, err := net.Listen("tcp", ":22300")
	if err != nil {
		log.Error().Msgf("listen failed, err:%v\n", err)
		return
	}
	log.Info().Msgf("server start successfully!")

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Error().Msgf("accept failed, err:%v\n", err)
			continue
		}
		go pkg.Process(&pkg.Connection{
			Conn:                     &conn,
			WriteLock:                sync.Mutex{},
			StreamToBackEndConnStore: common.NewRwmap(),
			StreamStore:              common.NewRwmap(),
		})
	}
}
