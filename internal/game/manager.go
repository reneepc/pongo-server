package game

import "sync"

type SessionManager struct {
	Sessions map[string]*GameSession
	sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		Sessions: make(map[string]*GameSession),
	}
}

func (sm *SessionManager) AddSession(id string, session *GameSession) {
	sm.Lock()
	defer sm.Unlock()

	sm.Sessions[id] = session
}

func (sm *SessionManager) RemoveSession(id string) {
	sm.Lock()
	defer sm.Unlock()

	delete(sm.Sessions, id)
}

func (sm *SessionManager) GetSession(id string) *GameSession {
	sm.Lock()
	defer sm.Unlock()

	return sm.Sessions[id]
}
