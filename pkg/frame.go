package pkg

import (
	"encoding/binary"
	"net"
)

type Frame struct {
	magic    byte
	hType    byte
	streamId uint32
	pLength  uint32
	payload  []byte
}

func (f *Frame) Write(conn *net.Conn) error {
	if _, err := (*conn).Write([]byte{f.magic}); err != nil {
		return err
	}
	if _, err := (*conn).Write([]byte{f.hType}); err != nil {
		return err
	}
	if err := binary.Write(*conn, binary.BigEndian, f.streamId); err != nil {
		return err
	}
	if err := binary.Write(*conn, binary.BigEndian, f.pLength); err != nil {
		return err
	}
	if _, err := (*conn).Write(f.payload); err != nil {
		return err
	}
	return nil
}
