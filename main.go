package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
)

const (
	HOST = "localhost"
	PORT = "8888"
	TYPE = "tcp"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	address := fmt.Sprintf("%s:%s", HOST, PORT)
	listener, err := net.Listen(TYPE, address)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	slog.Info("Server Listening", "address", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	slog.Info("New Client Connected", "addr", conn.RemoteAddr().String())

	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			slog.Info("Client Disconnected", "addr", conn.RemoteAddr().String())
			return
		}

		receivedData := string(buffer[:n])
		fmt.Printf("Received: %s\n", receivedData)

		conn.Write([]byte("Server says: " + receivedData))
	}
}
