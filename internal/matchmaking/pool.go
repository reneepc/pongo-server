package matchmaking

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/reneepc/pongo-server/internal/game"
)

type PlayerPool struct {
	sync.Mutex
	Players     []*game.Network
	matchSignal chan struct{}
}

func NewPlayerPool() *PlayerPool {
	pool := &PlayerPool{
		Players: make([]*game.Network, 0),
	}

	pool.matchSignal = make(chan struct{})

	go pool.StartMatchmaking()

	return pool
}

func (p *PlayerPool) AddPlayer(player *game.Network) {
	p.Lock()
	defer p.Unlock()

	player.JoinTime = time.Now()

	p.Players = append(p.Players, player)

	p.matchSignal <- struct{}{}
}

func (p *PlayerPool) RemovePlayer(player *game.Network) {
	p.Lock()
	defer p.Unlock()

	for i, poolPlayer := range p.Players {
		if poolPlayer == player {
			p.Players = append(p.Players[:i], p.Players[i+1:]...)
			return
		}
	}
}

func (p *PlayerPool) FindMatch() (*game.Network, *game.Network) {
	p.Lock()
	defer p.Unlock()

	if len(p.Players) < 2 {
		return nil, nil
	}

	p1, p2 := p.Players[0], p.Players[1]
	p.Players = p.Players[2:]

	return p1, p2
}

func (p *PlayerPool) StartMatchmaking() {
	for {
		<-p.matchSignal

		p1, p2 := p.FindMatch()
		if p1 == nil || p2 == nil {
			continue
		}

		startNewGameSession(p1, p2)
	}
}

func startNewGameSession(p1, p2 *game.Network) {
	go func() {
		for {
			select {
			case <-p1.Ctx.Done():
				slog.Warn("Player 1 disconnected", slog.String("name", p1.Name))
				p2.Cancel()
				return
			case <-p2.Ctx.Done():
				slog.Warn("Player 2 disconnected", slog.String("name", p2.Name))
				p1.Cancel()
				return
			case <-time.After(3 * time.Second):
				err := p1.Conn.WriteJSON(map[string]string{
					"message":       "Game is running",
					"ping":          p1.Latency.String(),
					"opponent_ping": p2.Latency.String(),
				})
				if err != nil {
					slog.Error("Failed to write to player", slog.String("name", p1.Name), slog.Any("error", err))
				}

				err = p2.Conn.WriteJSON(map[string]string{
					"message":       "Game is running",
					"ping":          p2.Latency.String(),
					"opponent_ping": p1.Latency.String(),
				})
				if err != nil {
					slog.Error("Failed to write to player", slog.String("name", p2.Name), slog.Any("error", err))
				}
			}
		}
	}()

	go broadcasterReader(p1, p2)
	go broadcasterReader(p2, p1)
}

func broadcasterReader(player *game.Network, opponent *game.Network) {
	for {
		select {
		case <-player.Ctx.Done():
			slog.Info("Stopping read for player", slog.String("name", player.Name))
			opponent.Cancel()
			return
		default:
			_, msg, err := player.Conn.ReadMessage()
			if err != nil {
				slog.Error("Error reading player input", slog.Any("error", err))
			}

			var input map[string]any
			if err := json.Unmarshal(msg, &input); err != nil {
				slog.Error("Invalid input from player", slog.Any("error", err))
				continue
			}

			slog.Info("Player input received", slog.Any("action", input), slog.String("player", player.Name))

			err = opponent.Conn.WriteJSON(map[string]any{
				"message": "Opponent moved",
				"input":   input,
			})
			if err != nil {
				slog.Error("Error sending input to opponent", slog.Any("error", err))
			}
		}
	}
}
