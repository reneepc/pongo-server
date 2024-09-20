package game

import (
	"log/slog"
	"time"

	"github.com/gandarez/pong-multiplayer-go/pkg/engine/ball"
	"github.com/gandarez/pong-multiplayer-go/pkg/engine/level"
	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
	"github.com/google/uuid"
)

const (
	MaxScore     = 10
	ScreenWidth  = 640
	ScreenHeight = 480
)

type GameSession struct {
	ID      string
	player1 *Player
	player2 *Player
	ball    *ball.Ball
	level   level.Level
	ticker  *time.Ticker
}

func NewGameSession(player1 *Player, player2 *Player) *GameSession {
	return &GameSession{
		ID:      uuid.NewString(),
		player1: player1,
		player2: player2,
		ball:    ball.New(geometry.Left, ScreenWidth, ScreenHeight, level.Medium),
		level:   level.Medium,
	}
}

func (session *GameSession) Start() {
	session.ticker = time.NewTicker(time.Second / 60)
	defer session.ticker.Stop()

	for {
		select {
		case <-session.player1.Network.Ctx.Done():
			session.handleDisconnection(session.player1)
			return
		case <-session.player2.Network.Ctx.Done():
			session.handleDisconnection(session.player2)
		case <-session.ticker.C:
			session.update()
			session.broadcastGameState()
		}
	}
}

func (session *GameSession) update() {
	session.player1.ProcessInputs()
	session.player2.ProcessInputs()

	session.ball.Update(session.player1.basePlayer.Bounds(), session.player2.basePlayer.Bounds())

	if scored, scorer := session.ball.CheckGoal(); scored {
		session.handleScore(scorer)
	}
}

func (session *GameSession) broadcastGameState() {
	player1 := PlayerState{
		PositionY: session.player1.basePlayer.Position().Y,
		Score:     session.player1.score,
		Side:      session.player1.side,
		Ping:      session.player1.Network.Latency,
	}

	player2 := PlayerState{
		PositionY: session.player2.basePlayer.Position().Y,
		Score:     session.player2.score,
		Side:      session.player2.side,
		Ping:      session.player2.Network.Latency,
	}

	session.player1.Network.Send(GameState{
		BallPosition: session.ball.Position(),
		Current:      player1,
		Opponent:     player2,
	})
	session.player2.Network.Send(GameState{
		BallPosition: session.ball.Position(),
		Current:      player2,
		Opponent:     player1,
	})
}

func (session *GameSession) handleDisconnection(disconnectedPlayer *Player) {
	disconnectedPlayer.Terminate()

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

func (session *GameSession) handleScore(scorer geometry.Side) {
	if session.player1.side == scorer {
		session.player1.score++
	} else {
		session.player2.score++
	}

	session.resetBall(scorer)

	if session.player1.score == MaxScore || session.player2.score == MaxScore {
		session.ticker.Stop()
		session.endGame()
	}
}

func (session *GameSession) endGame() {
	if session.player1.score == MaxScore {
		session.player1.Won()
		session.player2.Lost()
	} else {
		session.player2.Won()
		session.player1.Lost()
	}

	sessionManager.RemoveSession(session.ID)
}

func (session *GameSession) resetBall(scorer geometry.Side) {
	if scorer == geometry.Left {
		session.ball = session.ball.Reset(geometry.Right)
	} else {
		session.ball = session.ball.Reset(geometry.Left)
	}
}
