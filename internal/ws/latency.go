package ws

import (
	"log/slog"
	"time"

	"github.com/reneepc/pongo-server/internal/game"
)

func (s Server) measureLatency(player *game.Network) {
	handlePong(player)
	go sendPingMessages(player)
}

func sendPingMessages(player *game.Network) {
	for {
		select {
		case <-player.Ctx.Done():
			slog.Info("Stopping ping messages", slog.String("name", player.Name))
			return

		case <-time.After(5 * time.Second):
			player.Ping()
		}
	}
}

func handlePong(player *game.Network) {
	player.Conn.SetPongHandler(func(appData string) error {
		player.Latency = time.Since(player.LastPingTime)
		return nil
	})
}
