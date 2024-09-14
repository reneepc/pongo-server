package game

import (
	"time"

	"github.com/gandarez/pong-multiplayer-go/pkg/engine/ball"
	"github.com/gandarez/pong-multiplayer-go/pkg/engine/level"
	"github.com/gandarez/pong-multiplayer-go/pkg/engine/player"
)

type Player struct {
	BasePlayer player.Player
	Network    NetPlayer
	Score      int
}

type GameSession struct {
	Ball    ball.Ball
	Level   level.Level
	Player1 Player
	Player2 Player
}

func (session *GameSession) Start() {
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	for {
		select {
		// TODO: Handle players disconnection
		case <-session.Player1.Network.Ctx.Done():
			session.Player2.Network.Cancel()
			return
		case <-session.Player2.Network.Ctx.Done():
			session.Player1.Network.Cancel()
		case <-ticker.C:
			session.Update()
			session.BroadcastGameState()
		}
	}
}

func (session *GameSession) Update() {
}

func (session *GameSession) BroadcastGameState() {
}
