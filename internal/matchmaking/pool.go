package matchmaking

import (
	"sync"
	"time"

	"github.com/reneepc/pongo-server/internal/game/player"
)

const MaxWaitTime = 30 * time.Second

type PlayerPool struct {
	sync.Mutex
	Players []*player.Player
}

func NewPlayerPool() *PlayerPool {
	return &PlayerPool{
		Players: make([]*player.Player, 0),
	}
}

func (p *PlayerPool) AddPlayer(player *player.Player) {
	p.Lock()
	defer p.Unlock()

	player.JoinTime = time.Now()

	p.Players = append(p.Players, player)
}

func (p *PlayerPool) RemovePlayer(player *player.Player) {
	p.Lock()
	defer p.Unlock()

	for i, poolPlayer := range p.Players {
		if poolPlayer == player {
			p.Players = append(p.Players[:i], p.Players[i+1:]...)
			return
		}
	}
}

func (p *PlayerPool) FindMatch() (*player.Player, *player.Player) {
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
		p1, p2 := p.FindMatch()
		if p1 != nil && p2 != nil {
			startNewGameSession(p1, p2)
		}

		time.Sleep(1 * time.Second)
	}
}

func startNewGameSession(p1, p2 *player.Player) {
	go func() {
		time.Sleep(3 * time.Second)

		p1.Conn.WriteJSON(map[string]string{
			"message": "Game session started",
		})

		p2.Conn.WriteJSON(map[string]string{
			"message": "Game session started",
		})
	}()
}
