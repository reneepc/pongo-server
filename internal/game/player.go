package game

import (
	"context"
	"time"

	"github.com/gandarez/pong-multiplayer-go/pkg/engine/player"
	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
	"github.com/gorilla/websocket"
)

const defaultSpeed = 4

type Player struct {
	BasePlayer *player.Player
	Network    Network
	Score      int
	InputQueue chan PlayerInput
}

type Network struct {
	Conn         *websocket.Conn    `json:"-"`
	Name         string             `json:"name"`
	Latency      time.Duration      `json:"latency"`
	JoinTime     time.Time          `json:"-"`
	LastPingTime time.Time          `json:"-"`
	Ctx          context.Context    `json:"-"`
	Cancel       context.CancelFunc `json:"-"`
}

func NewPlayer(network *Network, side geometry.Side, screenWidth, screenHeight float64) *Player {
	basePlayer := player.New(network.Name, side, screenWidth, screenHeight, 10)

	player := &Player{
		BasePlayer: basePlayer,
		Network:    *network,
		Score:      0,
		InputQueue: make(chan PlayerInput, 100),
	}

	return player
}

func NewNetwork(conn *websocket.Conn, name string) *Network {
	ctx, cancel := context.WithCancel(context.Background())

	player := &Network{
		Conn:     conn,
		Name:     name,
		Latency:  0,
		JoinTime: time.Now(),
		Ctx:      ctx,
		Cancel:   cancel,
	}

	return player
}

func (p *Player) ProcessInputs() {
	for {
		select {
		case input := <-p.InputQueue:
			switch {
			case input.Up:
				p.MoveUp()
			case input.Down:
				p.MoveDown()
			}
		default:
			return
		}
	}
}

func (p *Player) MoveUp() {
	p.BasePlayer.SetPosition(p.BasePlayer.Position().Y - defaultSpeed)
}

func (p *Player) MoveDown() {
	p.BasePlayer.SetPosition(p.BasePlayer.Position().Y + defaultSpeed)
}
