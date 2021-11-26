package main

import (
	"../pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"sync"
	"../common"
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
			StreamStore:              make(map[uint32]*pkg.Stream),
		})
	}
}