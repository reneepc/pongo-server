package matchmaking

import (
	"sync"
	"time"

	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
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
	player1 := game.NewPlayer(p1, geometry.Left, game.ScreenWidth, game.ScreenHeight)
	player2 := game.NewPlayer(p2, geometry.Right, game.ScreenWidth, game.ScreenHeight)

	session := game.NewGameSession(player1, player2)

	game.GetSessionManager().AddSession(session.ID, session)

	go session.Start()

	player1.StartInputReader()
	player2.StartInputReader()
}
