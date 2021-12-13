package client

import (
	"bufio"
	"fmt"
	"github.com/ZijieSong/rrdp/common"
	"github.com/ZijieSong/rrdp/pkg"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

func Proxy(options *CliOptions) (err error) {
	//1. tcp connection to proxy server
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
	}
	go pkg.Process(proxyServer)
	log.Info().Msgf("proxy server connect success!")

	//2. pfctl nat
	if err := SetupNatRules(options); err != nil {
		return err
	}
	log.Info().Msgf("nat rules setup success!")

	//3. listen on localhost
	if err := ListenOnLocal(proxyServer); err != nil {
		return err
	}
	log.Info().Msgf("exit...")

	return nil
}

func ListenOnLocal(proxyServer *pkg.Connection) (err error) {
	ln, err := net.Listen("tcp", "127.0.0.1:12300")
	listen := ln.(*net.TCPListener)
	if err != nil {
		log.Error().Msgf("listen failed, err:%v\n", err)
		return
	}

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			log.Error().Msgf("accept failed, err:%v\n", err)
			continue
		}

		realDst, err := NatLookup(conn)
		log.Info().Msgf("accept new connection, and the real dest is %s", realDst)
		if err != nil {
			_ = conn.Close()
			log.Error().Msgf("cannot get real dest for %s", conn.RemoteAddr())
			continue
		}

		stream, err := pkg.CreateStream(realDst, proxyServer, conn)
		if err != nil {
			log.Error().Msgf("Create stream failed, %s", err.Error())
			continue
		}

		go func() {
			if _, err := io.Copy(stream, conn); err != nil {
				log.Error().Msgf("error copy from %s to stream: %s", stream.RealDest, err.Error())
			} else {
				log.Info().Msgf("tcp connection close, to %s", stream.RealDest)
			}
			_ = stream.Close()
		}()

	}
}

func SetupNatRules(options *CliOptions) error {
	//write to natrdr.pf
	natrdr := fmt.Sprintf("%s/natrdr.pf", options.ConfigPath)
	file, err := os.Create(natrdr)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)

	ports := arrayToString(options.LocalPorts.Value(), " ")
	if _, err = w.WriteString(fmt.Sprintf("rdr pass on lo0 inet proto tcp from any to 127.0.0.1 port { %s } -> 127.0.0.1 port 12300\n", ports)); err != nil {
		return err
	}
	if _, err = w.WriteString(fmt.Sprintf("pass out route-to lo0 inet proto tcp from any to 127.0.0.1 port { %s } keep state\n", ports)); err != nil {
		return err
	}
	if err = w.Flush(); err != nil {
		return err
	}

	//setup
	cmd := exec.Command("pfctl", "-f", natrdr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return err
	}

	return nil
}

func arrayToString(a []string, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

func NatLookup(c *net.TCPConn) (*net.TCPAddr, error) {
	const (
		PF_INOUT     = 0
		PF_IN        = 1
		PF_OUT       = 2
		IOC_OUT      = 0x40000000
		IOC_IN       = 0x80000000
		IOC_INOUT    = IOC_IN | IOC_OUT
		IOCPARM_MASK = 0x1FFF
		LEN          = 4*16 + 4*4 + 4*1
		// #define	_IOC(inout,group,num,len) (inout | ((len & IOCPARM_MASK) << 16) | ((group) << 8) | (num))
		// #define	_IOWR(g,n,t)	_IOC(IOC_INOUT,	(g), (n), sizeof(t))
		// #define DIOCNATLOOK		_IOWR('D', 23, struct pfioc_natlook)
		DIOCNATLOOK = IOC_INOUT | ((LEN & IOCPARM_MASK) << 16) | ('D' << 8) | 23
	)
	fd, err := syscall.Open("/dev/pf", 0, syscall.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer syscall.Close(fd)
	nl := struct { // struct pfioc_natlook
		saddr, daddr, rsaddr, rdaddr       [16]byte
		sxport, dxport, rsxport, rdxport   [4]byte
		af, proto, protoVariant, direction uint8
	}{
		af:        syscall.AF_INET,
		proto:     syscall.IPPROTO_TCP,
		direction: PF_OUT,
	}
	saddr := c.RemoteAddr().(*net.TCPAddr)
	daddr := c.LocalAddr().(*net.TCPAddr)
	copy(nl.saddr[:], saddr.IP)
	copy(nl.daddr[:], daddr.IP)
	nl.sxport[0], nl.sxport[1] = byte(saddr.Port>>8), byte(saddr.Port)
	nl.dxport[0], nl.dxport[1] = byte(daddr.Port>>8), byte(daddr.Port)
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), DIOCNATLOOK, uintptr(unsafe.Pointer(&nl))); errno != 0 {
		return nil, errno
	}
	var addr net.TCPAddr
	addr.IP = nl.rdaddr[:4]
	addr.Port = int(nl.rdxport[0])<<8 | int(nl.rdxport[1])
	return &addr, nil
}
