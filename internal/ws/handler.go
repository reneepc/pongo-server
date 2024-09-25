package ws

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/reneepc/pongo-server/internal/game"
	"github.com/reneepc/pongo-server/internal/matchmaking"
)

// Server is the WebSocket server
//
// It is responsible for handling incoming connections and managing the player pool.
type Server struct {
	PlayerPool *matchmaking.PlayerPool
	upgrader   websocket.Upgrader
}

func New() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		PlayerPool: matchmaking.NewPlayerPool(),
	}
}

func (s *Server) HandleConnections(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", slog.Any("error", err))
		return
	}

	// Wait for initial player info
	var info game.GameInfo
	if err := conn.ReadJSON(&info); err != nil {
		err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Failed to read player info"), time.Now().Add(time.Second))
		if err != nil {
			slog.Error("Failed to write close message after reading wrongly formatted player info", slog.Any("error", err))
		}
		slog.Error("Failed to read player info", slog.Any("error", err))
		return
	}

	slog.Info("New player connected", slog.String("name", info.PlayerName))

	newPlayer := game.NewNetwork(conn, info)

	// Starts ping measurement
	s.measureLatency(newPlayer)

	// Setup connection close handler and context cancellation on disconnect
	s.handleClosedConnection(newPlayer)

	s.PlayerPool.AddPlayer(newPlayer)
}

func (s *Server) handleClosedConnection(player *game.Network) {
	player.Conn.SetCloseHandler(func(code int, text string) error {
		slog.Info("Connection closed", slog.String("name", player.GameInfo.PlayerName), slog.Int("code", code), slog.String("text", text))
		player.Cancel()
		s.PlayerPool.RemovePlayer(player)
		return nil
	})
}
