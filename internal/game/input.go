package game

import (
	"encoding/json"
	"log/slog"
)

type PlayerInput struct {
	Up   bool `json:"up"`
	Down bool `json:"down"`
}

func inputReader(player *Player) {
	for {
		_, msg, err := player.Network.Conn.ReadMessage()
		if err != nil {
			slog.Error("Error reading player input", slog.Any("error", err))
			continue
		}

		var input PlayerInput
		if err := json.Unmarshal(msg, &input); err != nil {
			slog.Error("Failed to unmarshal input", slog.Any("error", err))
			continue
		}

		player.Input = input
	}
}
