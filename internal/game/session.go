package game

import (
	"log/slog"
	"time"

	"github.com/gandarez/pong-multiplayer-go/pkg/engine/ball"
	"github.com/gandarez/pong-multiplayer-go/pkg/engine/level"
	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
	"github.com/google/uuid"
)

// GameSession represents a multiplayer session between two players
//
// It contains physics related objects: the ball and the players, the
// game server clock (ticker), and selected game level.
//
// The ID identifies a session for the purposes of listing, streaming, and
// replaying a given session.
type GameSession struct {
	ID      string
	player1 *Player
	player2 *Player
	ball    ball.Ball
	level   level.Level
	ticker  *time.Ticker
}

func NewGameSession(player1 *Player, player2 *Player) *GameSession {
	return &GameSession{
		ID:      uuid.NewString(),
		player1: player1,
		player2: player2,
		ball:    ball.NewLocal(float64(player1.ScreenWidth), float64(player1.ScreenHeight), level.Medium),
		level:   level.Medium,
	}
}

// Start begins the game loop
//
// The game is processed in a fixed time step loop, given by the server clock (ticker).
// The game loop process the player inputs, updates the game physics, broadcasts the
// game state to the players.
//
// It also handles players disconnections, scores, and game ending.
func (session *GameSession) Start() {
	session.ticker = time.NewTicker(time.Second / 60)
	defer session.ticker.Stop()

	session.ready()

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

			if session.gameEnded() {
				session.ticker.Stop()
				session.endGame()
			}
		}
	}
}

func (session *GameSession) ready() {
	go session.player1.Network.Send(ReadyMessage{
		Ready:        true,
		Name:         session.player1.PlayerName,
		OpponentName: session.player2.PlayerName,
		Side:         session.player1.side,
		OpponentSide: session.player2.side,
	})
	go session.player2.Network.Send(ReadyMessage{
		Ready:        true,
		Name:         session.player2.PlayerName,
		OpponentName: session.player1.PlayerName,
		Side:         session.player2.side,
		OpponentSide: session.player1.side,
	})

	slog.Info("Game started", slog.String("session_id", session.ID), slog.Any("player1", session.player1), slog.Any("player2", session.player2))
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
		Name:      session.player1.PlayerName,
		PositionY: session.player1.basePlayer.Position().Y,
		Score:     session.player1.score,
		Side:      session.player1.side,
		Ping:      session.player1.Network.Latency.Milliseconds(),
		Winner:    session.winner(session.player1),
	}

	player2 := PlayerState{
		Name:      session.player2.PlayerName,
		PositionY: session.player2.basePlayer.Position().Y,
		Score:     session.player2.score,
		Side:      session.player2.side,
		Ping:      session.player2.Network.Latency.Milliseconds(),
		Winner:    session.winner(session.player2),
	}

	err := session.player1.Network.Send(GameState{
		Ball:     ballState(session.ball),
		Current:  player1,
		Opponent: player2,
	})
	if err != nil {
		slog.Error("Error sending game state to player 1", slog.Any("error", err), slog.Any("player", session.player1))
	}

	err = session.player2.Network.Send(GameState{
		Ball:     ballState(session.ball),
		Current:  player2,
		Opponent: player1,
	})
	if err != nil {
		slog.Error("Error sending game state to player 2", slog.Any("error", err), slog.Any("player", session.player2))
	}
}

func (session *GameSession) handleDisconnection(disconnectedPlayer *Player) {
	disconnectedPlayer.Terminate()

	var remainingPlayer *Player
	if disconnectedPlayer == session.player1 {
		remainingPlayer = session.player2
	} else {
		remainingPlayer = session.player1
	}

	slog.Warn("Player disconnected", slog.String("name", disconnectedPlayer.Network.GameInfo.PlayerName))

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
}

func (session *GameSession) gameEnded() bool {
	return session.winner(session.player1) || session.winner(session.player2)
}

func (session *GameSession) endGame() {
	if session.player1.score == session.player1.MaxScore {
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
		session.ball = session.ball.Reset()
	} else {
		session.ball = session.ball.Reset()
	}
}

func (session *GameSession) winner(player *Player) bool {
	return player.score >= player.MaxScore
}
