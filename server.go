package link

import (
	"net"
	"time"
)

type Server struct {
	manager      *Manager
	listener     net.Listener
	protocol     Protocol
	sendChanSize int
}

type Handler interface {
	HandleSession(*Session)
}

var _ Handler = HandlerFunc(nil)

type HandlerFunc func(*Session)

func (f HandlerFunc) HandleSession(session *Session) {
	f(session)
}

func NewServer(l net.Listener, p Protocol, sendChanSize int) *Server {
	return &Server{
		manager:      NewManager(),
		listener:     l,
		protocol:     p,
		sendChanSize: sendChanSize,
	}
}

func (server *Server) Listener() net.Listener {
	return server.listener
}

func (server *Server) Serve(handler Handler) error {
	var tempDelay time.Duration

	for {
		conn, err := server.listener.Accept()
		if err != nil {

			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}

			return err
		}

		go func() {
			codec, err := server.protocol.NewCodec(conn)
			if err != nil {
				conn.Close()
				return
			}
			session := server.manager.NewSession(codec, server.sendChanSize)
			handler.HandleSession(session)
		}()
	}
}

func (server *Server) GetSession(sessionID uint64) *Session {
	return server.manager.GetSession(sessionID)
}

func (server *Server) Stop() {
	server.listener.Close()
	server.manager.Dispose()
}
