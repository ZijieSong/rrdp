package client

import (
	"github.com/ZijieSong/rrdp/common"
	"github.com/ZijieSong/rrdp/pkg"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
)

func Expose(options *CliOptions) error {
	wg := sync.WaitGroup{}
	ports := options.ExposedPorts.Value()
	wg.Add(len(ports))

	for _, port := range ports {

		// per port per tcp conn
		go func(port string) {
			defer func() {
				wg.Done()
			}()
			conn, err := net.Dial("tcp", options.RemoteServer)
			if err != nil {
				return
			}
			defer conn.Close()
			proxyServer := &pkg.Connection{
				Conn:                     &conn,
				WriteLock:                sync.Mutex{},
				StreamToBackEndConnStore: common.NewRwmap(),
				StreamStore:              common.NewRwmap(),
				NextStreamIdLock:         sync.Mutex{},
				NextStreamId:             0x0,
				BackendStoreLock:         sync.RWMutex{},
				ExposedPort:              port,
			}
			log.Info().Msgf("proxy server connect success!")

			streamId := proxyServer.GetNextStreamId()
			portExposedHandshakeStream := &pkg.Stream{
				StreamId: streamId,
				Conn:     proxyServer,
			}
			err = portExposedHandshakeStream.SendPortExposedRequest(port)
			if err != nil {
				log.Error().Msgf("error send port exposed request: ", err.Error())
				return
			}

			pkg.Process(proxyServer)
		}(port)
	}

	//wait all goroutine end
	wg.Wait()
	log.Info().Msgf("finished port exposed, quit...")
	return nil
}
