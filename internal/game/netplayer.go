package game

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

type NetPlayer struct {
	Conn         *websocket.Conn    `json:"-"`
	Name         string             `json:"name"`
	Latency      time.Duration      `json:"latency"`
	JoinTime     time.Time          `json:"-"`
	LastPingTime time.Time          `json:"-"`
	Ctx          context.Context    `json:"-"`
	Cancel       context.CancelFunc `json:"-"`
}

func NewNetPlayer(conn *websocket.Conn, name string) *NetPlayer {
	ctx, cancel := context.WithCancel(context.Background())

	player := &NetPlayer{
		Conn:     conn,
		Name:     name,
		Latency:  0,
		JoinTime: time.Now(),
		Ctx:      ctx,
		Cancel:   cancel,
	}

	return player
}
