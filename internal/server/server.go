package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/reneepc/pongo-server/internal/ws"
)

type Server struct {
	httpServer *http.Server
}

func New() *Server {
	httpServer := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
	}
}

func (s *Server) Start(addr string, wsServer *ws.Server) error {
	http.HandleFunc("/multiplayer", wsServer.HandleConnections)

	s.httpServer.Addr = addr

	slog.Info("Server started", slog.String("addr", addr))
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}
