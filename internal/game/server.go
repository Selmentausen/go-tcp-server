package game

import (
	"fmt"
	"go-tcp-server/internal/domain"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
)

type Server struct {
	clients map[net.Conn]*domain.Player
	mu      sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		clients: make(map[net.Conn]*domain.Player),
	}
}

func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			slog.Error("Failed to close server properly", "err", err)
			os.Exit(1)
		}
	}(listener)

	slog.Info("Server listening", "address", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error("Failed to close connection", "error", err, "conn", conn)
		}
	}(conn)

	// Handshake
	_, err := conn.Write([]byte("MSG:Enter your name:\n"))
	if err != nil {
		slog.Error("Failed to write message", "error", err)
	}
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return
	}
	name := strings.TrimSpace(string(buffer[:n]))

	// Create player
	s.mu.Lock()
	colorIdx := len(s.clients) % len(domain.COLORS)
	player := domain.NewPlayer(
		name,
		rand.Intn(domain.MAP_WIDTH),
		rand.Intn(domain.MAP_HEIGHT),
		domain.COLORS[colorIdx],
	)
	s.clients[conn] = player
	s.mu.Unlock()

	go s.writePump(conn, player)

	slog.Info("Player joined", "name", player.Name)
	s.broadcastGameView()
	s.broadcastMessage(fmt.Sprintf("--- %s joined ---\n", name), player, 0)

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
			fullMessage := fmt.Sprintf("%s: %s\n", player.Name, msg)
			s.broadcastMessage(fullMessage, player, domain.CHAT_RADIUS)
		}
	}

}

// writePump listens to the channel and write to the TCP socket
func (s *Server) writePump(conn net.Conn, player *domain.Player) {
	for msg := range player.MsgChan {
		_, err := conn.Write([]byte(msg))
		if err != nil {
			break
		}
	}
}

func (s *Server) broadcastGameView() {
	view := s.renderView()
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.clients {
		select {
		case p.MsgChan <- "MAP:" + view:
		default:
		}
	}
}

func (s *Server) broadcastMessage(message string, sender *domain.Player, radius int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.clients {
		if p == sender || isNearby(sender, p, radius) {
			select {
			case p.MsgChan <- "MSG:" + message:
			default:
			}
		}
	}
}

func (s *Server) removeClient(conn net.Conn) {
	s.mu.Lock()
	player := s.clients[conn]
	delete(s.clients, conn)
	s.mu.Unlock()

	if player != nil {
		close(player.MsgChan) // Stop the writePump
		slog.Info("player Disconnected", "name", player.Name)
		s.broadcastMessage(fmt.Sprintf("--- %s left the chat ---\n", player.Name), nil, 0)
		go s.broadcastGameView()
	}
}

func (s *Server) handleCommand(sender net.Conn, command string, player *domain.Player) {
	switch command {
	case "/w":
		if player.Y < domain.MAP_HEIGHT {
			player.Y++
		}
	case "/s":
		if player.Y > 0 {
			player.Y--
		}
	case "/d":
		if player.X < domain.MAP_WIDTH {
			player.X++
		}
	case "/a":
		if player.X > 0 {
			player.X--
		}
	default:
		_, err := sender.Write([]byte("MSG:Unknown command. Use /w /a /s /d to move.\n"))
		if err != nil {
			slog.Error("Failed to send command", "error", err)
			return
		}
		return
	}
	_, _ = sender.Write([]byte(fmt.Sprintf("MSG:You moved to (%d,%d)\n", player.X, player.Y)))
	s.broadcastGameView()
}

func (s *Server) renderView() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var sb strings.Builder

	for y := domain.MAP_HEIGHT; y >= 0; y-- {
		for x := 0; x < domain.MAP_WIDTH; x++ {
			playerHere := ""
			for _, p := range s.clients {
				if p.X == x && p.Y == y {
					if len(p.Name) > 0 {
						playerHere = p.Color + string(p.Name[0]) + domain.RESET
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
	return sb.String()
}

func isNearby(p1, p2 *domain.Player, radius int) bool {
	if radius == 0 {
		return true
	}
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return (dx*dx + dy*dy) <= (radius * radius)
}
