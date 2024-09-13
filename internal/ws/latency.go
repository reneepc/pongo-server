package ws

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/reneepc/pongo-server/internal/game/player"
)

func (s Server) measureLatency(player *player.Player) {
	handlePong(player)
	go sendPingMessages(player)
}

func sendPingMessages(player *player.Player) {
	for {
		select {
		case <-player.Ctx.Done():
			slog.Info("Stopping ping messages", slog.String("name", player.Name))
			return

		case <-time.After(5 * time.Second):
			player.LastPingTime = time.Now()

			if err := player.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Failed to send ping message", slog.Any("error", err))
				return
			}
		}
	}
}

func handlePong(player *player.Player) {
	player.Conn.SetPongHandler(func(appData string) error {
		player.Latency = time.Since(player.LastPingTime)
		return nil
	})
}
