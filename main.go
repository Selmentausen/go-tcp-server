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

type Player struct {
	Name  string
	X     int
	Y     int
	Color string
}

type Server struct {
	clients map[net.Conn]*Player
	mu      sync.Mutex
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	server := &Server{
		clients: make(map[net.Conn]*Player),
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

	player := &Player{
		Name: strings.TrimSpace(string(buffer[:n])),
		X:    0,
		Y:    0,
	}

	s.mu.Lock()
	s.clients[conn] = player
	s.mu.Unlock()

	slog.Info("Player joined", "name", player.Name)
	s.broadcastMessage(fmt.Sprintf("--- %s joined at (0, 0) ---\n", player.Name), conn)

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

		if strings.HasPrefix(msg, "/") {
			s.handleCommand(conn, msg, player)
		} else {
			fullMessage := fmt.Sprintf("[%s]: %s\n", player.Name, msg)
			s.broadcastMessage(fullMessage, conn)
		}
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
	player := s.clients[conn]
	delete(s.clients, conn)
	s.mu.Unlock()

	slog.Info("player Disconnected", "name", player.Name)
	s.broadcastMessage(fmt.Sprintf("--- %s left the chat ---\n", player.Name), nil)
}

func (s *Server) handleCommand(sender net.Conn, command string, player *Player) {
	switch command {
	case "/w":
		player.Y++
	case "/s":
		player.Y--
	case "/a":
		player.X++
	case "/d":
		player.X--
	default:
		sender.Write([]byte("Unknown command. Use /w /a /s /d to move.\n"))
		return
	}
	sender.Write([]byte(fmt.Sprintf("You moved to (%d,%d)\n", player.X, player.Y)))
}
