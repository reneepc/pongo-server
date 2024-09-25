package game

import "github.com/gandarez/pong-multiplayer-go/pkg/geometry"

// ReadyMessage is the first message sent to the players after connecting and finding a match
//
// It conveys information about the opponent player name and which side each player is allocated.
type ReadyMessage struct {
	Ready        bool          `json:"ready"`
	Name         string        `json:"name"`
	OpponentName string        `json:"opponent_name"`
	Side         geometry.Side `json:"side"`
	OpponentSide geometry.Side `json:"opponent_side"`
}
