package game

import (
	"log/slog"
	"time"

	"github.com/gandarez/pong-multiplayer-go/pkg/engine/player"
	"github.com/gandarez/pong-multiplayer-go/pkg/geometry"
	"github.com/gorilla/websocket"
)

const defaultSpeed = 4

// Player unifies all multiplayer concerns about a player in the game
//
// The physics concerns are handled by the basePlayer, which shares the local
// update logic with the client's local player processing.
//
// The inputQueue streamlines the input processing, allowing the game loop to
// process the player inputs in a controlled manner.
type Player struct {
	*Network
	basePlayer player.Player
	side       geometry.Side
	score      int8
	inputQueue chan PlayerInput
}

func NewPlayer(network *Network, side geometry.Side) *Player {
	basePlayer := player.NewLocal(network.PlayerName, side, float64(network.ScreenWidth), float64(network.ScreenHeight), float64(network.FieldBorderWidth))

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
			p.basePlayer.Update(player.Input{
				Up:   input.Up,
				Down: input.Down,
			})
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
