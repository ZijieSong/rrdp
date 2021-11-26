package pkg

import (
	"../common"
	"encoding/binary"
	"github.com/rs/zerolog/log"
	"io"
	"net"
)

func Process(conn *Connection) {
	defer conn.Close()

	for {

		//read header
		header := make([]byte, common.HEADER_SIZE)
		_, err := io.ReadFull(*conn.Conn, header)
		if err != nil {
			log.Error().Msgf("socket read header error: %s", err.Error())
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
		_, err = io.ReadFull(*conn.Conn, payload)
		if err != nil {
			log.Error().Msgf("socket read data(len=%d) error: %s", pLength, err.Error())
			return
		}
		log.Info().Msgf("streamId is %d, pLength is %d ", streamId, pLength)

		//handle payload
		switch hType {
		case common.HANDLE_TYPE_SYN:
			{
				realDest := string(payload)
				stream := &Stream{
					StreamId: streamId,
					Conn:     conn,
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
						_ = stream.Close()
					}
				}()
				log.Info().Msgf("register from stream %s to conn %s", streamId, backendConn.RemoteAddr())
			}
		case common.HANDLE_TYPE_CHAT:
			{
				backend := conn.StreamToBackEndConnStore.Get(streamId).(*net.TCPConn)
				if backend != nil {
					_, err := (*backend).Write(payload)
					if err != nil {
						log.Error().Msgf("Error write to %s, error is %s", (*backend).RemoteAddr(), err.Error())
						stream := conn.StreamStore[streamId]
						if stream != nil {
							_ = stream.Close()
						}
					}
					log.Info().Msgf("send msg to stream %s, remote %s", streamId, backend.RemoteAddr())
				} else {
					//has not sync
					stream := &Stream{
						StreamId: streamId,
						Conn:     conn,
					}
					_ = stream.SendFinalToStreamPeer()
				}
			}
		case common.HANDLE_TYPE_FINAL:
			{
				stream := conn.StreamStore[streamId]
				if stream != nil {
					stream.DeRegisterToConn()
				}
			}
			log.Info().Msgf("final send to stream %s", streamId)
		}
	}
}
