package game

// AddSpectator adds a spectator to a given session from which it will
// receive buffered game updates
func (session *GameSession) AddSpectator(spectator *Network) {
	session.spectatorMutex.Lock()
	defer session.spectatorMutex.Unlock()

	session.spectators = append(session.spectators, spectator)
}

// RemoveSpectator removes a spectator from a given session
func (session *GameSession) RemoveSpectator(spectator *Network) {
	session.spectatorMutex.Lock()
	defer session.spectatorMutex.Unlock()

	for i, s := range session.spectators {
		if s == spectator {
			session.spectators = append(session.spectators[:i], session.spectators[i+1:]...)
			break
		}
	}
}
