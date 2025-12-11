package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
)

var COLORS = []string{
	"\033[31m", // Red
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[34m", // Blue
	"\033[35m", // Magenta
	"\033[36m", // Cyan
}

const RESET = "\033[0m"

const (
	HOST        = "localhost"
	PORT        = "8888"
	TYPE        = "tcp"
	MAP_WIDTH   = 20
	MAP_HEIGHT  = 10
	CHAT_RADIUS = 5
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

	x := rand.Intn(MAP_WIDTH)
	y := rand.Intn(MAP_HEIGHT)
	s.mu.Lock()
	colorIdx := len(s.clients) % len(COLORS)
	s.mu.Unlock()
	player := &Player{
		Name:  strings.TrimSpace(string(buffer[:n])),
		X:     x,
		Y:     y,
		Color: COLORS[colorIdx],
	}

	s.mu.Lock()
	s.clients[conn] = player
	s.mu.Unlock()

	slog.Info("Player joined", "name", player.Name)
	s.broadcastMessage(fmt.Sprintf("--- %s joined at (%d, %d) ---\n", player.Name, x, y), player, 0)

	conn.Write([]byte(s.renderView()))

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
			s.broadcastMessage(fullMessage, player, CHAT_RADIUS)
		}
	}
}

func (s *Server) broadcastMessage(message string, sender *Player, radius int) {
	s.mu.Lock()
	defer s.mu.Unlock() // Ensure unlock happens when function exits

	for conn, p := range s.clients {
		if p != sender {
			if sender == nil || isNearby(sender, p, radius) {
				conn.Write([]byte(message))
			}
		}
	}
}

func (s *Server) removeClient(conn net.Conn) {
	s.mu.Lock()
	player := s.clients[conn]
	delete(s.clients, conn)
	s.mu.Unlock()

	slog.Info("player Disconnected", "name", player.Name)
	s.broadcastMessage(fmt.Sprintf("--- %s left the chat ---\n", player.Name), nil, 0)
}

func (s *Server) handleCommand(sender net.Conn, command string, player *Player) {
	switch command {
	case "/w":
		player.Y++
	case "/s":
		player.Y--
	case "/d":
		player.X++
	case "/a":
		player.X--
	default:
		sender.Write([]byte("Unknown command. Use /w /a /s /d to move.\n"))
		return
	}
	sender.Write([]byte(fmt.Sprintf("You moved to (%d,%d)\n", player.X, player.Y)))
	s.broadcastGameView()
}

func (s *Server) renderView() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var sb strings.Builder

	sb.WriteString("\033[H\033[2J")
	sb.WriteString("=== GOPHER ARENA ===\n")

	for y := MAP_HEIGHT; y >= 0; y-- {
		for x := 0; x < MAP_WIDTH; x++ {
			playerHere := ""
			for _, p := range s.clients {
				if p.X == x && p.Y == y {
					if len(p.Name) > 0 {
						playerHere = p.Color + string(p.Name[0]) + RESET
					} else {
						playerHere = "P"
					}
					break
				}
			}
			if playerHere != "" {
				sb.WriteString(playerHere)
			} else {
				sb.WriteString(".")
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\nUse /w /a /s /d to move. Chat to talk.\n")
	return sb.String()
}

func (s *Server) broadcastGameView() {
	view := s.renderView()

	s.mu.Lock()
	defer s.mu.Unlock()

	for conn, _ := range s.clients {
		conn.Write([]byte(view))
	}
}

func isNearby(p1, p2 *Player, radius int) bool {
	if radius == 0 {
		return true
	}

	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return (dx*dx + dy*dy) <= (radius * radius)
}
