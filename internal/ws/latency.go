package ws

import (
	"log/slog"
	"time"

	"github.com/reneepc/pongo-server/internal/game"
)

func (s Server) measureLatency(player *game.Network) {
	// Set the receiver of client's response to the ping message.
	handlePong(player)

	// Send periodically ping messages to the player's client.
	go sendPingMessages(player)
}

func sendPingMessages(player *game.Network) {
	for {
		select {
		case <-player.Ctx.Done():
			slog.Info("Stopping ping messages", slog.String("name", player.GameInfo.PlayerName))
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
