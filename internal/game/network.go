package game

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Network struct {
	Conn         *websocket.Conn    `json:"-"`
	mutex        sync.Mutex         `json:"-"`
	closed       bool               `json:"-"`
	Latency      time.Duration      `json:"latency"`
	JoinTime     time.Time          `json:"-"`
	LastPingTime time.Time          `json:"-"`
	Ctx          context.Context    `json:"-"`
	Cancel       context.CancelFunc `json:"-"`
	PlayerInfo
}

func NewNetwork(conn *websocket.Conn, playerInfo PlayerInfo) *Network {
	ctx, cancel := context.WithCancel(context.Background())

	player := &Network{
		Conn:       conn,
		Latency:    0,
		JoinTime:   time.Now(),
		Ctx:        ctx,
		Cancel:     cancel,
		PlayerInfo: playerInfo,
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
		slog.Error("Error while sending ping to player", slog.Any("error", err), slog.String("name", n.PlayerInfo.Name))
	}
}
