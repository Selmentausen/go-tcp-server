package main

import (
	"go-tcp-server/internal/game"
	"log/slog"
	"os"
)

const (
	HOST = "localhost"
	PORT = "8888"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	server := game.NewServer()
	address := HOST + ":" + PORT
	if err := server.Start(address); err != nil {
		slog.Error("Server crashed", "error", err)
		os.Exit(1)
	}
}
