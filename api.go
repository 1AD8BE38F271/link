package link

import (
	"io"
	"net"
)

type Dialer interface {
	Dial() (net.Conn, error)
}

type DialerFunc func() (net.Conn, error)

func (f DialerFunc) Dial() (net.Conn, error) {
	return f()
}

type Protocol interface {
	NewCodec(rw io.ReadWriter) (Codec, error)
}

type Codec interface {
	Receive() (interface{}, error)
	Send(interface{}) error
	Close() error
}

type Handler interface {
	HandleSession(*Session)
}

type HandlerFunc func(*Session)

func (f HandlerFunc) HandleSession(session *Session) {
	f(session)
}

func CreateCodec(dialer Dialer, protocol Protocol) (Codec, error) {
	conn, err := dialer.Dial()
	if err != nil {
		return nil, err
	}

	codec, err := protocol.NewCodec(conn)
	if err != nil {
		return nil, err
	}
	return codec, nil
}

func CreateSession(conn net.Conn, protocol Protocol, sendChanSize int) (*Session, error) {
	codec, err := protocol.NewCodec(conn)
	if err != nil {
		return nil, err
	}
	return NewSession(codec, sendChanSize), nil
}
