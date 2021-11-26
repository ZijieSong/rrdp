package pkg

import (
	"../common"
	"github.com/rs/zerolog/log"
	"net"
)

type Stream struct {
	StreamId uint32
	Conn     *Connection
	RealDest string
}

// WriteData writes data to stream, sending a dataframe per call
func (s *Stream) WriteData(hType byte, data []byte) error {
	frame := &Frame{
		magic:    common.MAGIC,
		hType:    hType,
		streamId: s.StreamId,
		pLength:  uint32(len(data)),
		payload:  data,
	}
	s.Conn.WriteLock.Lock()
	defer s.Conn.WriteLock.Unlock()
	return frame.Write(s.Conn.Conn)
}

func (s *Stream) Close() error {
	s.DeRegisterToConn()
	return s.SendFinalToStreamPeer()
}

func (s *Stream) ShakeHands() error {
	return s.WriteData(common.HANDLE_TYPE_SYN, []byte(s.RealDest))
}

func (s *Stream) DeRegisterToConn() {
	//close tcp peer
	backend := s.Conn.StreamToBackEndConnStore.Get(s.StreamId)
	if backend != nil {
		back := backend.(*net.TCPConn)
		if err := (*back).Close(); err != nil {
			log.Error().Msgf("close to %s error: %s", (*back).RemoteAddr(), err.Error())
		}
	}
	//rm from map
	s.Conn.StreamToBackEndConnStore.Delete(s.StreamId)
	delete(s.Conn.StreamStore, s.StreamId)
}

func (s *Stream) SendFinalToStreamPeer() error {
	return s.WriteData(common.HANDLE_TYPE_FINAL, []byte("bye"))
}

func (s *Stream) RegisterToConn(backendConn *net.TCPConn) {
	s.Conn.StreamToBackEndConnStore.Put(s.StreamId, backendConn)
	s.Conn.StreamStore[s.StreamId] = s
}

func (s *Stream) Write(bytes []byte) (int, error) {
	err := s.WriteData(common.HANDLE_TYPE_CHAT, bytes)
	return len(bytes), err
}
