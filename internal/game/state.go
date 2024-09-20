package game

import (
	"time"

	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
)

type GameState struct {
	BallPosition geometry.Vector `json:"ball_position"`
	Current      PlayerState     `json:"current"`
	Opponent     PlayerState     `json:"opponent"`
}

type PlayerState struct {
	Position geometry.Vector `json:"position"`
	Side     geometry.Side   `json:"side"`
	Score    int             `json:"score"`
	Ping     time.Duration   `json:"ping"`
}
