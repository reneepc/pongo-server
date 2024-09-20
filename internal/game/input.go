package game

import (
	"log/slog"
)

type PlayerInput struct {
	Up   bool `json:"up"`
	Down bool `json:"down"`
}

func (player *Player) StartInputReader() {
	go func() {
		for {
			select {
			case <-player.Network.Ctx.Done():
				return
			default:
				var input PlayerInput
				if err := player.Network.Conn.ReadJSON(&input); err != nil {
					slog.Error("Error reading player input", slog.Any("error", err))
					continue
				}

				slog.Info("Received input", slog.Any("input", input))

				player.inputQueue <- input
			}
		}
	}()
}
