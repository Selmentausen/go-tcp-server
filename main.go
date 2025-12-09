package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	HOST = "localhost"
	PORT = "8888"
	TYPE = "tcp"
)

type Server struct {
	clients map[net.Conn]string
	mu      sync.Mutex
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	server := &Server{
		clients: make(map[net.Conn]string),
	}

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
		go server.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("Enter your name:\n"))

	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil {
		return
	}

	name := strings.TrimSpace(string(buffer[:n]))

	s.mu.Lock()
	s.clients[conn] = name
	s.mu.Unlock()

	slog.Info("Client joined", "name", name, "addr", conn.RemoteAddr().String())
	s.broadcastMessage(fmt.Sprintf("--- %s joined the chat ---\n", name), conn)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			s.removeClient(conn)
			return
		}

		msg := strings.TrimSpace(string(buffer[:n]))
		if msg == "" {
			continue
		}
		fullMessage := fmt.Sprintf("[%s]: %s\n", name, msg)
		s.broadcastMessage(fullMessage, conn)
	}
}

func (s *Server) broadcastMessage(message string, sender net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock() // Ensure unlock happens when function exits

	for conn, _ := range s.clients {
		if conn != sender {
			conn.Write([]byte(message))
		}
	}
}

func (s *Server) removeClient(conn net.Conn) {
	s.mu.Lock()
	name := s.clients[conn]
	delete(s.clients, conn)
	s.mu.Unlock()

	slog.Info("Client Disconnected", "name", name)
	s.broadcastMessage(fmt.Sprintf("--- %s left the chat ---\n", name), nil)
}
