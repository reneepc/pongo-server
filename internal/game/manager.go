package game

import "sync"

var sessionManager = NewSessionManager()

// SessionManager stores all active game sessions
//
// This struct is supposed to be a singleton
type SessionManager struct {
	Sessions map[string]*GameSession
	sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		Sessions: make(map[string]*GameSession),
	}
}

// GetSessionManager returns the singleton instance of the SessionManager
func GetSessionManager() *SessionManager {
	return sessionManager
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

func (sm *SessionManager) Session(id string) *GameSession {
	sm.Lock()
	defer sm.Unlock()

	return sm.Sessions[id]
}

func (sm *SessionManager) GetSessions() []*GameSession {
	sm.Lock()
	defer sm.Unlock()

	sessions := make([]*GameSession, 0, len(sm.Sessions))
	for _, session := range sm.Sessions {
		sessions = append(sessions, session)
	}
	return sessions
}
