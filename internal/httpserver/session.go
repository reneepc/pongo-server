package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/reneepc/pongo-server/internal/game"
)

type SessionInfo struct {
	ID      string `json:"id"`
	Player1 string `json:"player1"`
	Player2 string `json:"player2"`
}

// handleSessions returns a list of active sessions
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	sessions := game.GetSessionManager().GetSessions()
	sessionList := make([]SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		sessionList = append(sessionList, SessionInfo{
			ID:      session.ID,
			Player1: session.Player1.PlayerName,
			Player2: session.Player2.PlayerName,
		})
	}

	json.NewEncoder(w).Encode(sessionList)
}
