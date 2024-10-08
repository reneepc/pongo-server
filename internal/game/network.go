package game

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Network stores a player's websocket related information
type Network struct {
	Conn         *websocket.Conn    `json:"-"`
	mutex        sync.Mutex         `json:"-"`
	closed       bool               `json:"-"`
	Latency      time.Duration      `json:"latency"`
	JoinTime     time.Time          `json:"-"`
	LastPingTime time.Time          `json:"-"`
	Ctx          context.Context    `json:"-"`
	Cancel       context.CancelFunc `json:"-"`
	GameInfo
}

func NewNetwork(conn *websocket.Conn, info GameInfo) *Network {
	ctx, cancel := context.WithCancel(context.Background())

	player := &Network{
		Conn:     conn,
		Latency:  0,
		JoinTime: time.Now(),
		Ctx:      ctx,
		Cancel:   cancel,
		GameInfo: info,
	}

	return player
}

// Send is responsible for marshalling and sending data to the player's client
func (n *Network) Send(data any) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if err := n.Conn.WriteJSON(data); err != nil {
		slog.Error("Error writing to player", slog.Any("error", err))
		return err
	}

	return nil
}

// Terminate is responsible for closing the player's connection and canceling the connection's context
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

// Ping is responsible for sending messages to measure the latency between the server and the player
func (n *Network) Ping() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.LastPingTime = time.Now()

	if err := n.Conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
		slog.Error("Error while sending ping to player", slog.Any("error", err), slog.String("name", n.GameInfo.PlayerName))
	}
}

func (n *Network) opponentDisconnect() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.Conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Opponent disconnected"), time.Now().Add(time.Second))
}
