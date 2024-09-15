package game

import (
	"encoding/json"
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
				_, msg, err := player.Network.Conn.ReadMessage()
				if err != nil {
					slog.Error("Error reading player input", slog.Any("error", err))
					continue
				}

				var input PlayerInput
				if err := json.Unmarshal(msg, &input); err != nil {
					slog.Error("Error unmarshalling player input", slog.Any("error", err))
					continue
				}

				if input.Time.IsZero() {
					input.Time = time.Now()
				}

				player.InputQueue <- input
			}
		}
	}()
}
