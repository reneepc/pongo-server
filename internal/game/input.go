package game

import (
	"log/slog"
)

// PlayerInput stores the player's input
//
// It's supposed to be received from the client only when there is
// an effective action from the player (up or down)
type PlayerInput struct {
	Up   bool `json:"up"`
	Down bool `json:"down"`
}

func (player *Player) StartInputReader() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Recovered from panic", slog.Any("error", r))
			}

			slog.Error("Player input reader stopped", slog.Any("player", player))
			player.Terminate()
		}()

		for {
			select {
			case <-player.Network.Ctx.Done():
				return
			default:
				var input PlayerInput
				if err := player.Network.Conn.ReadJSON(&input); err != nil {
					slog.Error("Error reading player input", slog.Any("error", err))
					player.Terminate()

					return
				}

				if !input.Up && !input.Down {
					continue
				}

				slog.Info("Received input", slog.Any("input", input))

				player.inputQueue <- input
			}
		}
	}()
}
