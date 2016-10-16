package link

import (
	"io"
	"net"
	"time"
)

type Protocol interface {
	NewCodec(rw io.ReadWriter) (Codec, error)
}

type Codec interface {
	Receive() (interface{}, error)
	Send(interface{}) error
	Close() error
}

func CreateServer(network, address string, protocol Protocol, sendChanSize int) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, protocol, sendChanSize), nil
}

func CreateSession(conn net.Conn, protocol Protocol, sendChanSize int) (*Session, error) {
	codec, err := protocol.NewCodec(conn)
	if err != nil {
		return nil, err
	}
	return NewSession(codec, sendChanSize), nil
}

func ConnectTimeout(network, address string, timeout time.Duration, protocol Protocol, sendChanSize int) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	codec, err := protocol.NewCodec(conn)
	if err != nil {
		return nil, err
	}
	return NewSession(codec, sendChanSize), nil
}
