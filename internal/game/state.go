package game

import (
	"github.com/gandarez/pong-multiplayer-go/pkg/engine/ball"
	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
)

// GameState is a snapshot of the game physics at a given time.
//
// The game state is used to synchronize the game between the server and the clients.
// At a constant rate, the server sends state updates to the clients in response to the client's inputs.
// The clients use the state updates to render the game and predict the game physics.
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
	Name      string        `json:"name"`
	PositionY float64       `json:"position_y"`
	Side      geometry.Side `json:"side"`
	Score     int8          `json:"score"`
	Ping      int64         `json:"ping"`
	Winner    bool          `json:"winner,omitempty"`
}

func ballState(ball ball.Ball) BallState {
	return BallState{
		Angle:    ball.Angle(),
		Bounces:  ball.Bounces(),
		Position: ball.Position(),
	}
}
