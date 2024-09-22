package main

import (
	"context"
	"flag"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/reneepc/pongo-server/internal/game"
)

func main() {
	serverAddr := flag.String("server", "localhost:8080", "The WebSocket server address")
	playerName := flag.String("name", "TestPlayer", "The player's name")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	u := url.URL{Scheme: "ws", Host: *serverAddr, Path: "/multiplayer"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		slog.Error("Failed to connect to WebSocket server", slog.Any("error", err))
		return
	}
	defer conn.Close()

	conn.SetPingHandler(func(appData string) error {
		if err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second)); err != nil {
			slog.Error("Failed to send pong message", slog.Any("error", err))
		}
		return nil
	})

	playerInfo := game.PlayerInfo{Name: *playerName}
	if err := conn.WriteJSON(playerInfo); err != nil {
		slog.Error("Failed to send player info", slog.Any("error", err))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleServerMessages(ctx, conn)
	handleInterrupt(conn, cancel)
}

func handleServerMessages(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var gameState game.GameState
			if err := conn.ReadJSON(&gameState); err != nil {
				slog.Info("Connection closed or read error", slog.Any("error", err))
				return
			}

			slog.Info("Received game state", slog.Any("state", gameState))
		}
	}
}

func handleInterrupt(conn *websocket.Conn, cancel context.CancelFunc) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	slog.Info("Interrupt received, closing connection", slog.Any("signal", <-interrupt))
	cancel()

	err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Client closed connection"), time.Now().Add(time.Second))
	if err != nil {
		slog.Error("Error sending close message", slog.Any("error", err))
		return
	}
}
