package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/reneepc/pongo-server/internal/server"
	"github.com/reneepc/pongo-server/internal/ws"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithCancel(context.Background())

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	slog.Info("Starting Pong Multiplayer Server", slog.String("port", port))

	httpServer := server.New()
	wsServer := ws.New()

	go func() {
		host := fmt.Sprintf("0.0.0.0:%s", port)
		err := httpServer.Start(host, wsServer)
		if err != nil {
			slog.Error("Error starting server", slog.Any("error", err))
			cancel()

			return
		}
	}()

	shutdown(ctx, httpServer, wsServer)
}

func shutdown(ctx context.Context, s *server.Server, wsServer *ws.Server) {
	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quitSignal:
	case <-ctx.Done():
	}

	slog.Info("Shutting down server")
	for _, player := range wsServer.PlayerPool.Players {
		player.Terminate()
	}

	if err := s.Shutdown(); err != nil {
		slog.Error("Error shutting down server", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Server shut down gracefully")
}
