package ws

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/reneepc/pongo-server/internal/game"
	"github.com/reneepc/pongo-server/internal/matchmaking"
)

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
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", slog.Any("error", err))
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		slog.Error("Failed to read message", slog.Any("error", err))
		return
	}

	playerInfo, err := PlayerInfoFromMsg(msg)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		slog.Error("Invalid player info", slog.Any("error", err))
		return
	}

	slog.Info("New player connected", slog.String("name", playerInfo.Name))

	newPlayer := game.NewNetwork(conn, playerInfo.Name)

	s.measureLatency(newPlayer)
	s.handleClosedConnection(newPlayer)

	s.PlayerPool.AddPlayer(newPlayer)
}

func (s *Server) handleClosedConnection(player *game.Network) {
	player.Conn.SetCloseHandler(func(code int, text string) error {
		slog.Info("Connection closed", slog.String("name", player.Name))
		player.Cancel()
		s.PlayerPool.RemovePlayer(player)
		return nil
	})
}
