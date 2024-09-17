package game

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gandarez/pong-multiplayer-go/pkg/engine/player"
	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
	"github.com/gorilla/websocket"
)

const defaultSpeed = 4

type Player struct {
	basePlayer *player.Player
	Network    *Network
	side       geometry.Side
	score      int
	inputQueue chan PlayerInput
}

type Network struct {
	Conn         *websocket.Conn    `json:"-"`
	mutex        sync.Mutex         `json:"-"`
	closed       bool               `json:"-"`
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
		basePlayer: basePlayer,
		Network:    network,
		side:       side,
		score:      0,
		inputQueue: make(chan PlayerInput, 100),
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

func (n *Network) opponentDisconnect() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.Conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Opponent disconnected"), time.Now().Add(time.Second))
}

func (n *Network) Send(data any) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if err := n.Conn.WriteJSON(data); err != nil {
		slog.Error("Error writing to player", slog.Any("error", err))
	}
}

func (n *Network) Terminate() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if n.Conn != nil && !n.closed {
		if err := n.Conn.Close(); err != nil {
			slog.Warn("Failed to close connection", slog.Any("error", err))
		}
		n.closed = true
	}

	n.Cancel()
}

func (n *Network) Ping() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.LastPingTime = time.Now()

	if err := n.Conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
		slog.Error("Error while sending ping to player", slog.Any("error", err), slog.String("name", n.Name))
	}
}

func (p *Player) Won() {
	p.Network.mutex.Lock()
	defer p.Network.mutex.Unlock()

	if err := p.Network.Conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "You won!"), time.Now().Add(time.Second)); err != nil {
		slog.Error("Error writing to player", slog.Any("error", err))
	}

	p.Network.Terminate()
}

func (p *Player) Lost() {
	p.Network.mutex.Lock()
	defer p.Network.mutex.Unlock()

	if err := p.Network.Conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "You lost!"), time.Now().Add(time.Second)); err != nil {
		slog.Error("Error writing to player", slog.Any("error", err))
	}

	p.Network.Terminate()
}

func (p *Player) ProcessInputs() {
	for {
		select {
		case input := <-p.inputQueue:
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
	p.basePlayer.SetPosition(p.basePlayer.Position().Y - defaultSpeed)
}

func (p *Player) MoveDown() {
	p.basePlayer.SetPosition(p.basePlayer.Position().Y + defaultSpeed)
}

func (p *Player) Terminate() {
	p.Network.Terminate()
}
