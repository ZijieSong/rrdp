package pkg

import (
	"../common"
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"strings"
)

func Process(peerConn *Connection) {
	ch := make(chan struct{},1)
	defer func() {
		peerConn.Close()
		ch <- struct{}{}
	}()

	for {

		//read header
		header := make([]byte, common.HEADER_SIZE)
		_, err := io.ReadFull(*peerConn.Conn, header)
		if err != nil {
			log.Info().Msgf("peer conn closed...")
			return
		}

		//1.magic 1byte
		magic := header[0]
		if magic != common.MAGIC {
			log.Error().Msgf("socket read header magic error: %x", magic)
			return
		}
		//2.type 1byte
		hType := header[1]
		//3.streamId 4byte
		streamId := binary.BigEndian.Uint32(header[2:6])
		//4.payload length
		pLength := binary.BigEndian.Uint32(header[6:10])

		//read payload
		payload := make([]byte, pLength)
		_, err = io.ReadFull(*peerConn.Conn, payload)
		if err != nil {
			log.Error().Msgf("socket read data(len=%d) error: %s", pLength, err.Error())
			return
		}
		log.Debug().Msgf("streamId is %d, pLength is %d ", streamId, pLength)

		//handle payload
		switch hType {
		case common.HANDLE_TYPE_EXPOSE_REQ:
			{
				go func() {
					portExposedHandshakeStream := &Stream{
						StreamId: streamId,
						Conn:     peerConn,
					}

					exposePortPeer := strings.Split(string(payload), ":")
					if len(exposePortPeer) != 2 {
						_ = portExposedHandshakeStream.SendPortExposedResult(false)
						return
					}
					localPort := exposePortPeer[1]
					peerPort := exposePortPeer[0]

					ln, err := net.Listen("tcp", fmt.Sprintf(":%s", localPort))
					if err != nil {
						log.Error().Msgf("listen failed, err:%v\n", err)
						_ = portExposedHandshakeStream.SendPortExposedResult(false)
						return
					}
					listen := ln.(*net.TCPListener)
					go func() {
						<-ch
						_ = listen.Close()
						log.Info().Msgf("listen socket closed: %s", listen.Addr())
					}()
					if err = portExposedHandshakeStream.SendPortExposedResult(true); err != nil {
						//peer already closed, just close goroutine
						log.Error().Msgf("send port exposed result failed, err: ", err.Error())
						return
					}
					log.Info().Msgf("listen socket open: %s", listen.Addr())

					for {
						clientConn, err := listen.AcceptTCP()
						if err != nil {
							log.Error().Msgf("accept failed, err:%v\n", err)
							return
						}
						log.Info().Msgf("accept new connection, and the client is %s", clientConn.RemoteAddr())

						realAddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%s", peerPort))
						stream, err := CreateStream(realAddr, peerConn, clientConn)
						if err != nil {
							log.Error().Msgf("Create chat stream failed, %s", err.Error())
							continue
						}

						go func() {
							if _, err := io.Copy(stream, clientConn); err != nil {
								log.Error().Msgf("error copy from %s to chat stream: %s", stream.RealDest, err.Error())
							} else {
								log.Info().Msgf("client tcp connection close, to %s", stream.RealDest)
							}
							_ = stream.Close()
						}()
					}
				}()
			}
		case common.HANDLE_TYPE_EXPOSE_RES:
			{
				if payload[0] == common.FALSE {
					log.Error().Msgf("cannot expose port %s to remote, please check if the port has already been listened on remote instance", peerConn.ExposedPort)
					return
				} else if payload[0] == common.TRUE {
					log.Info().Msgf("expose port %s to remote successfully", peerConn.ExposedPort)
				}
			}
		case common.HANDLE_TYPE_SYN:
			{
				realDest := string(payload)
				stream := &Stream{
					StreamId: streamId,
					Conn:     peerConn,
					RealDest: realDest,
				}
				backendConn, err := net.Dial("tcp", realDest)
				if err != nil {
					log.Error().Msgf("Error connecting to %s, error is :%s", realDest, err.Error())
					_ = stream.Close()
					continue
				}
				stream.RegisterToConn(backendConn.(*net.TCPConn))
				go func() {
					if _, err := io.Copy(stream, backendConn); err != nil {
						log.Error().Msgf("error copy from %s to stream: %s", stream.RealDest, err.Error())
					} else {
						log.Info().Msgf("peerConn closed, from %s to stream", stream.RealDest)
					}
					_ = stream.Close()
				}()
				log.Info().Msgf("register from stream %d to peerConn %s", streamId, backendConn.RemoteAddr())
			}
		case common.HANDLE_TYPE_CHAT:
			{
				backend := peerConn.StreamToBackEndConnStore.Get(streamId)
				if backend != nil {
					back := backend.(*net.TCPConn)
					_, err := (*back).Write(payload)
					if err != nil {
						log.Error().Msgf("Error write to %s, error is %s", (*back).RemoteAddr(), err.Error())
						stream := peerConn.StreamStore[streamId]
						if stream != nil {
							_ = stream.Close()
						}
					}
					log.Debug().Msgf("send msg to stream %d, remote %s", streamId, back.RemoteAddr())
				} else {
					//has not sync
					stream := &Stream{
						StreamId: streamId,
						Conn:     peerConn,
					}
					_ = stream.SendFinalToStreamPeer()
				}
			}
		case common.HANDLE_TYPE_FINAL:
			{
				stream := peerConn.StreamStore[streamId]
				if stream != nil {
					stream.DeRegisterToConn()
				}
			}
			log.Info().Msgf("final send to stream %d", streamId)
		}
	}
}
