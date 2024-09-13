package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/reneepc/pongo-server/internal/server"
	"github.com/reneepc/pongo-server/internal/ws"
)

func main() {
	port := flag.String("port", "8080", "The port on which to run the server")

	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	slog.Info("Starting Pong Multiplyaer Server", slog.String("port", *port))

	httpServer := server.New()
	wsServer := ws.New()

	go func() {
		host := fmt.Sprintf("0.0.0.0:%s", *port)
		err := httpServer.Start(host, wsServer)
		if err != nil {
			slog.Error("Error starting server", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	shutdown(httpServer, wsServer)
}

func shutdown(s *server.Server, wsServer *ws.Server) {
	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)

	<-quitSignal

	slog.Info("Shutting down server")
	for _, player := range wsServer.PlayerPool.Players {
		player.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server shutting down"))
		player.Conn.Close()
	}

	if err := s.Shutdown(); err != nil {
		slog.Error("Error shutting down server", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Server shut down gracefully")
}
