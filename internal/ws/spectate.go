package ws

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/reneepc/pongo-server/internal/game"
)

type SpectateRequest struct {
	SessionID string `json:"session_id"`
}

// HandleSpectatorConnections handles incoming spectator connections for a given session ID
//
// The spectator connection is upgraded to a websocket connection, and the session is retrieved
// from the session manager. If the session is not found, the connection is closed.
//
// The spectator is added to the session, and a close handler is set to remove the spectator
// from the session when the connection is closed.
func (s *Server) HandleSpectatorConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade spectator connection", slog.Any("error", err))
		return
	}

	var spectateRequest SpectateRequest
	if err := conn.ReadJSON(&spectateRequest); err != nil {
		err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Failed to read spectate request"), time.Now().Add(time.Second))
		if err != nil {
			slog.Error("Failed to write close message after reading wrongly formatted spectate request", slog.Any("error", err))
		}
		slog.Error("Failed to read spectate request", slog.Any("error", err))
		return
	}

	session := game.GetSessionManager().Session(spectateRequest.SessionID)
	if session == nil {
		err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Session not found"), time.Now().Add(time.Second))
		if err != nil {
			slog.Error("Failed to write close message after session not found", slog.Any("error", err))
		}
		slog.Error("Session not found", slog.String("session_id", spectateRequest.SessionID))
		return
	}

	spectator := game.NewNetwork(conn, game.GameInfo{})
	session.AddSpectator(spectator)

	s.handleSpectatorDisconnection(spectator, session)
}

func (s *Server) handleSpectatorDisconnection(spectator *game.Network, session *game.GameSession) {
	spectator.Conn.SetCloseHandler(func(code int, text string) error {
		slog.Info("Spectator connection closed", slog.Int("code", code), slog.String("text", text))
		session.RemoveSpectator(spectator)
		spectator.Cancel()
		return nil
	})
}
