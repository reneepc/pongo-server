package game

import (
	"log/slog"

	"github.com/gorilla/websocket"
)

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

					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						player.Terminate()
						return
					}

					continue
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
