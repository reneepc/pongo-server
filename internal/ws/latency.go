package ws

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/reneepc/pongo-server/internal/game"
)

func (s Server) measureLatency(player *game.NetPlayer) {
	handlePong(player)
	go sendPingMessages(player)
}

func sendPingMessages(player *game.NetPlayer) {
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

func handlePong(player *game.NetPlayer) {
	player.Conn.SetPongHandler(func(appData string) error {
		slog.Info("Received pong message", slog.String("message", appData))
		player.Latency = time.Since(player.LastPingTime)
		return nil
	})
}
