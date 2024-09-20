package game

import (
	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
)

type GameState struct {
	BallPosition geometry.Vector `json:"ball_position"`
	Current      PlayerState     `json:"current"`
	Opponent     PlayerState     `json:"opponent"`
}

type PlayerState struct {
	PositionY float64       `json:"position_y"`
	Side      geometry.Side `json:"side"`
	Score     int8          `json:"score"`
	Ping      int64         `json:"ping"`
}
