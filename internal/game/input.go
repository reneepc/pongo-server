package game

import (
	"log/slog"
	"time"
)

type PlayerInput struct {
	Up   bool      `json:"up"`
	Down bool      `json:"down"`
	Time time.Time `json:"time"`
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

				if !input.Up && !input.Down {
					continue
				}

				if input.Time.IsZero() {
					input.Time = time.Now()
				}

				player.inputQueue <- input
			}
		}
	}()
}
