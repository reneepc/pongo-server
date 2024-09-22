package game

import (
	"log/slog"
	"time"

	"github.com/gandarez/pong-multiplayer-go/pkg/engine/player"
	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
	"github.com/gorilla/websocket"
)

const defaultSpeed = 4

type Player struct {
	basePlayer player.Player
	Network    *Network
	side       geometry.Side
	score      int8
	inputQueue chan PlayerInput
}

func NewPlayer(network *Network, side geometry.Side) *Player {
	basePlayer := player.NewNetwork(network.PlayerInfo.Name, side, float64(network.PlayerInfo.ScreenWidth), float64(network.PlayerInfo.ScreenHeight))

	player := &Player{
		basePlayer: basePlayer,
		Network:    network,
		side:       side,
		score:      0,
		inputQueue: make(chan PlayerInput, 100),
	}

	return player
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
