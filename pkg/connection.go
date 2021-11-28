package pkg

import (
	"../common"
	"net"
	"sync"
)

type Connection struct {
	Conn                     *net.Conn
	WriteLock                sync.Mutex
	StreamToBackEndConnStore *common.Rwmap
	BackendStoreLock         sync.RWMutex
	StreamStore              map[uint32]*Stream
	NextStreamIdLock         sync.Mutex
	NextStreamId             uint32
	ExposedPort              string
}

func (c *Connection) Close() error {
	return (*c.Conn).Close()
}

func (c *Connection) GetNextStreamId() uint32 {
	c.NextStreamIdLock.Lock()
	defer c.NextStreamIdLock.Unlock()

	sid := c.NextStreamId
	if sid == 0xffffffff {
		c.NextStreamId = 0
	} else {
		c.NextStreamId = sid + 1
	}

	return sid
}
