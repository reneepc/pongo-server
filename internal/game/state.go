package game

import (
	"github.com/gandarez/pong-multiplayer-go/pkg/engine/ball"
	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
)

type GameState struct {
	Ball     BallState   `json:"ball"`
	Current  PlayerState `json:"current"`
	Opponent PlayerState `json:"opponent"`
}

type BallState struct {
	Angle    float64         `json:"angle"`
	Bounces  int             `json:"bounces"`
	Position geometry.Vector `json:"position"`
}

type PlayerState struct {
	PositionY float64       `json:"position_y"`
	Side      geometry.Side `json:"side"`
	Score     int8          `json:"score"`
	Ping      int64         `json:"ping"`
}

func ballState(ball ball.Ball) BallState {
	return BallState{
		Angle:    ball.Angle(),
		Bounces:  ball.Bounces(),
		Position: ball.Position(),
	}
}
