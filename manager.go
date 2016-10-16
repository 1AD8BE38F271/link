package link

import "sync"

const sessionMapNum = 32

type Manager struct {
	sessionMaps [sessionMapNum]sessionMap
	disposeFlag bool
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

type sessionMap struct {
	sync.RWMutex
	sessions map[uint64]*Session
}

func NewManager() *Manager {
	manager := &Manager{}
	for i := 0; i < len(manager.sessionMaps); i++ {
		manager.sessionMaps[i].sessions = make(map[uint64]*Session)
	}
	return manager
}

func (manager *Manager) Dispose() {
	manager.disposeOnce.Do(func() {
		manager.disposeFlag = true
		for i := 0; i < sessionMapNum; i++ {
			smap := &manager.sessionMaps[i]
			smap.Lock()
			for _, session := range smap.sessions {
				session.Close()
			}
			smap.Unlock()
		}
		manager.disposeWait.Wait()
	})
}

func (manager *Manager) NewSession(codec Codec, sendChanSize int) *Session {
	session := NewSession(codec, sendChanSize)
	session.addCloseCallback(manager.delSession)
	manager.putSession(session)
	return session
}

func (manager *Manager) GetSession(sessionID uint64) *Session {
	smap := &manager.sessionMaps[sessionID%sessionMapNum]
	smap.RLock()
	defer smap.RUnlock()

	session, _ := smap.sessions[sessionID]
	return session
}

func (manager *Manager) GetSessions() (sess []*Session) {
	for i := 0; i < sessionMapNum; i++ {
		smap := &manager.sessionMaps[i]
		smap.RLock()
		for _, session := range smap.sessions {
			sess = append(sess, session)
		}
		smap.RUnlock()
	}

	return
}

func (manager *Manager) GetSize() (size int) {
	for i := 0; i < sessionMapNum; i++ {
		smap := &manager.sessionMaps[i]
		smap.RLock()
		size += len(smap.sessions)
		smap.RUnlock()
	}
	return
}

func (manager *Manager) putSession(session *Session) {
	smap := &manager.sessionMaps[session.id%sessionMapNum]
	smap.Lock()
	defer smap.Unlock()

	smap.sessions[session.id] = session
	manager.disposeWait.Add(1)
}

func (manager *Manager) delSession(session *Session) {
	if manager.disposeFlag {
		//avoid deadload
		manager.disposeWait.Done()
		return
	}

	smap := &manager.sessionMaps[session.id%sessionMapNum]
	smap.Lock()
	defer smap.Unlock()

	delete(smap.sessions, session.id)
	manager.disposeWait.Done()
}
