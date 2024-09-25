package game

import (
	"log/slog"
	"sync"
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
	Player1 *Player
	Player2 *Player
	ball    ball.Ball
	level   level.Level
	ticker  *time.Ticker

	// Spectate
	spectators     []*Network
	spectatorMutex sync.Mutex
	stateBuffer    []GameState
	delayFrames    int
	bufferSize     int
}

func NewGameSession(player1 *Player, player2 *Player) *GameSession {
	return &GameSession{
		ID:      uuid.NewString(),
		Player1: player1,
		Player2: player2,
		ball:    ball.NewLocal(float64(player1.ScreenWidth), float64(player1.ScreenHeight), level.Medium),
		level:   level.Medium,

		// Spectate
		bufferSize:  60,
		delayFrames: 30,
		stateBuffer: make([]GameState, 0, 60),
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
		case <-session.Player1.Network.Ctx.Done():
			session.handleDisconnection(session.Player1)
			return
		case <-session.Player2.Network.Ctx.Done():
			session.handleDisconnection(session.Player2)
		case <-session.ticker.C:
			session.update()

			session.broadcastGameState()

			session.broadcastToSpectators()

			if session.gameEnded() {
				session.ticker.Stop()
				session.endGame()
			}
		}
	}
}

func (session *GameSession) ready() {
	go session.Player1.Network.Send(ReadyMessage{
		Ready:        true,
		Name:         session.Player1.PlayerName,
		OpponentName: session.Player2.PlayerName,
		Side:         session.Player1.side,
		OpponentSide: session.Player2.side,
	})
	go session.Player2.Network.Send(ReadyMessage{
		Ready:        true,
		Name:         session.Player2.PlayerName,
		OpponentName: session.Player1.PlayerName,
		Side:         session.Player2.side,
		OpponentSide: session.Player1.side,
	})

	slog.Info("Game started", slog.String("session_id", session.ID), slog.Any("player1", session.Player1), slog.Any("player2", session.Player2))
}

func (session *GameSession) update() {
	session.Player1.ProcessInputs()
	session.Player2.ProcessInputs()

	session.ball.Update(session.Player1.basePlayer.Bounds(), session.Player2.basePlayer.Bounds())

	if scored, goalSide := session.ball.CheckGoal(); scored {
		session.handleScore(goalSide)
	}

	// Buffer game state for streaming
	session.stateBuffer = append(session.stateBuffer, session.currentGameState())
	if len(session.stateBuffer) > session.bufferSize {
		session.stateBuffer = session.stateBuffer[1:]
	}
}

func (session *GameSession) broadcastGameState() {
	player1 := PlayerState{
		Name:      session.Player1.PlayerName,
		PositionY: session.Player1.basePlayer.Position().Y,
		Score:     session.Player1.score,
		Side:      session.Player1.side,
		Ping:      session.Player1.Network.Latency.Milliseconds(),
		Winner:    session.winner(session.Player1),
	}

	player2 := PlayerState{
		Name:      session.Player2.PlayerName,
		PositionY: session.Player2.basePlayer.Position().Y,
		Score:     session.Player2.score,
		Side:      session.Player2.side,
		Ping:      session.Player2.Network.Latency.Milliseconds(),
		Winner:    session.winner(session.Player2),
	}

	err := session.Player1.Network.Send(GameState{
		Ball:     ballState(session.ball),
		Current:  player1,
		Opponent: player2,
	})
	if err != nil {
		slog.Error("Error sending game state to player 1", slog.Any("error", err), slog.Any("player", session.Player1))
	}

	err = session.Player2.Network.Send(GameState{
		Ball:     ballState(session.ball),
		Current:  player2,
		Opponent: player1,
	})
	if err != nil {
		slog.Error("Error sending game state to player 2", slog.Any("error", err), slog.Any("player", session.Player2))
	}
}

func (session *GameSession) handleDisconnection(disconnectedPlayer *Player) {
	disconnectedPlayer.Terminate()

	var remainingPlayer *Player
	if disconnectedPlayer == session.Player1 {
		remainingPlayer = session.Player2
	} else {
		remainingPlayer = session.Player1
	}

	slog.Warn("Player disconnected", slog.String("name", disconnectedPlayer.Network.GameInfo.PlayerName))

	remainingPlayer.Network.opponentDisconnect()
	remainingPlayer.Terminate()
}

func (session *GameSession) handleScore(goalSide geometry.Side) {
	if session.Player1.side == goalSide {
		session.Player2.score++
	} else {
		session.Player1.score++
	}

	session.resetBall(goalSide)
}

func (session *GameSession) gameEnded() bool {
	return session.winner(session.Player1) || session.winner(session.Player2)
}

func (session *GameSession) endGame() {
	if session.Player1.score == session.Player1.MaxScore {
		session.Player1.Won()
		session.Player2.Lost()
	} else {
		session.Player2.Won()
		session.Player1.Lost()
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

func (session *GameSession) currentGameState() GameState {
	player1State := PlayerState{
		Name:      session.Player1.PlayerName,
		PositionY: session.Player1.basePlayer.Position().Y,
		Score:     session.Player1.score,
		Side:      session.Player1.side,
		Ping:      session.Player1.Network.Latency.Milliseconds(),
		Winner:    session.winner(session.Player1),
	}

	player2State := PlayerState{
		Name:      session.Player2.PlayerName,
		PositionY: session.Player2.basePlayer.Position().Y,
		Score:     session.Player2.score,
		Side:      session.Player2.side,
		Ping:      session.Player2.Network.Latency.Milliseconds(),
		Winner:    session.winner(session.Player2),
	}

	return GameState{
		Ball:     ballState(session.ball),
		Current:  player1State,
		Opponent: player2State,
	}
}

func (session *GameSession) broadcastToSpectators() {
	session.spectatorMutex.Lock()
	defer session.spectatorMutex.Unlock()

	if len(session.stateBuffer) == 0 {
		return
	}

	if len(session.stateBuffer) < session.delayFrames {
		return
	}

	delayedState := session.stateBuffer[len(session.stateBuffer)-session.delayFrames]

	for _, spectator := range session.spectators {
		spectator.Send(delayedState)
	}
}
