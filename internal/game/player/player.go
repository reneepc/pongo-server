package player

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	Conn         *websocket.Conn    `json:"-"`
	Name         string             `json:"name"`
	Latency      time.Duration      `json:"latency"`
	JoinTime     time.Time          `json:"-"`
	LastPingTime time.Time          `json:"-"`
	Ctx          context.Context    `json:"-"`
	Cancel       context.CancelFunc `json:"-"`
}

func NewPlayer(conn *websocket.Conn, name string) *Player {
	ctx, cancel := context.WithCancel(context.Background())

	player := &Player{
		Conn:     conn,
		Name:     name,
		Latency:  0,
		JoinTime: time.Now(),
		Ctx:      ctx,
		Cancel:   cancel,
	}

	return player
}
