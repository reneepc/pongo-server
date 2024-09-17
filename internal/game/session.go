package game

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gandarez/pong-multiplayer-go/pkg/engine/ball"
	"github.com/gandarez/pong-multiplayer-go/pkg/engine/level"
	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
	"github.com/gorilla/websocket"
)

type GameSession struct {
	Ball    *ball.Ball
	Level   *level.Level
	Player1 *Player
	Player2 *Player
	Ticker  *time.Ticker
}

func (session *GameSession) Start() {
	session.Ticker = time.NewTicker(time.Second / 60)
	defer session.Ticker.Stop()

	for {
		select {
		case <-session.Player1.Network.Ctx.Done():
			session.Player2.Network.Cancel()
			return
		case <-session.Player2.Network.Ctx.Done():
			session.Player1.Network.Cancel()
		case <-session.Ticker.C:
			session.Update()
			session.BroadcastGameState()
		}
	}
}

func (session *GameSession) Update() {
	session.Player1.ProcessInputs()
	session.Player2.ProcessInputs()

	session.Ball.Update(session.Player1.BasePlayer.Bounds(), session.Player2.BasePlayer.Bounds())

	if scored, scorer := session.Ball.CheckGoal(); scored {
		session.HandleScore(scorer)
	}
}

func (session *GameSession) BroadcastGameState() {
	state := GameState{
		BallPosition: session.Ball.Position(),
		Player1: PlayerState{
			Position: session.Player1.BasePlayer.Position(),
			Score:    session.Player1.Score,
			Ping:     session.Player1.Network.Latency,
		},
		Player2: PlayerState{
			Position: session.Player2.BasePlayer.Position(),
			Score:    session.Player2.Score,
			Ping:     session.Player2.Network.Latency,
		},
	}

	message, err := json.Marshal(state)
	if err != nil {
		slog.Error("Error marshalling game state", slog.Any("error", err))
		return
	}

	session.sendToPlayer(session.Player1, message)
	session.sendToPlayer(session.Player2, message)
}

func (session *GameSession) handleDisconnection(disconnectedPlayer *Player) {
	disconnectedPlayer.Terminate()

func (session *GameSession) HandleDisconnection(disconnectedPlayer *Player) {
	var remainingPlayer *Player
	if disconnectedPlayer == session.player1 {
		remainingPlayer = session.player2
	} else {
		remainingPlayer = session.player1
	}

	slog.Warn("Player disconnected", slog.String("name", disconnectedPlayer.Network.Name))

	remainingPlayer.Network.opponentDisconnect()
	remainingPlayer.Terminate()
}

func (session *GameSession) HandleScore(scorer geometry.Side) {
	if scorer == geometry.Left {
		session.Player1.Score++
	} else {
		session.Player2.Score++
	}

	session.Ball.Reset(scorer)
}
