package link

import (
	"container/list"
	"errors"
	"sync"
	"sync/atomic"
)

var SessionClosedError = errors.New("Session Closed")
var SessionBlockedError = errors.New("Session Blocked")

var globalSessionId uint64

type Session struct {
	id       uint64
	codec    Codec
	sendChan chan interface{}

	closeFlag      int32
	closeChan      chan int

	closeMutex     sync.Mutex
	closeCallbacks *list.List

}

func NewSession(codec Codec, sendChanSize int) *Session {
	session := &Session{
		codec:     codec,
		closeChan: make(chan int),
		id:        atomic.AddUint64(&globalSessionId, 1),
	}
	if sendChanSize > 0 {
		session.sendChan = make(chan interface{}, sendChanSize)
		go session.sendLoop()
	}
	return session
}

func (session *Session) ID() uint64 {
	return session.id
}

func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) == 1
}

func (session *Session) Close() error {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		err := session.codec.Close()
		close(session.closeChan)
		session.invokeCloseCallbacks()
		return err
	}
	return SessionClosedError
}

func (session *Session) Codec() Codec {
	return session.codec
}

func (session *Session) Receive() (interface{}, error) {
	msg, err := session.codec.Receive()
	if err != nil {
		session.Close()
	}
	return msg, err
}

func (session *Session) sendLoop() {
	defer session.Close()
	for {
		select {
		case msg, ok := <-session.sendChan:
			if !ok || session.codec.Send(msg) != nil {
				return
			}
		case <-session.closeChan:
			return
		}
	}
}

func (session *Session) Send(msg interface{}) (err error) {
	if session.IsClosed() {
		return SessionClosedError
	}
	if session.sendChan == nil {
		return session.codec.Send(msg)
	}
	select {
	case session.sendChan <- msg:
		return nil
	default:
		return SessionBlockedError
	}
}


type closeCallbackFunc func(*Session)

func (session *Session) addCloseCallback(callback closeCallbackFunc) {
	if session.IsClosed() {
		return
	}

	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	if session.closeCallbacks == nil {
		session.closeCallbacks = list.New()
	}

	session.closeCallbacks.PushBack(callback)
}

func (session *Session) invokeCloseCallbacks() {
	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	if session.closeCallbacks == nil {
		return
	}

	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		callback := i.Value.(closeCallbackFunc)
		callback(session)
	}
}
