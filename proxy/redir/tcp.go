package redir

import (
	"net"

	"github.com/ClashrAuto/Clashr/adapters/inbound"
	C "github.com/ClashrAuto/Clashr/constant"
	"github.com/ClashrAuto/Clashr/log"
	"github.com/ClashrAuto/Clashr/tunnel"
)

var (
	tun = tunnel.Instance()
)

type RedirListener struct {
	net.Listener
	address string
	closed  bool
}

func NewRedirProxy(addr string) (*RedirListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	rl := &RedirListener{l, addr, false}

	go func() {
		log.Infoln("Redir proxy listening at: %s", addr)
		for {
			c, err := l.Accept()
			if err != nil {
				if rl.closed {
					break
				}
				continue
			}
			go handleRedir(c)
		}
	}()

	return rl, nil
}

func (l *RedirListener) Close() {
	l.closed = true
	_ = l.Listener.Close()
}

func (l *RedirListener) Address() string {
	return l.address
}

func handleRedir(conn net.Conn) {
	target, err := parserPacket(conn)
	if err != nil {
		_ = conn.Close()
		return
	}
	_ = conn.(*net.TCPConn).SetKeepAlive(true)
	tun.Add(adapters.NewSocket(target, conn, C.REDIR, C.TCP))
}
