# Go TCP Game Server

A multiplayer terminal-based game engine built using raw TCP sockets in Go.

This project explores network programming and concurrency without using HTTP frameworks. It implements a custom text protocol to handle real-time movement and chat simultaneously.

## How it Works
*   **Architecture:** The server maintains persistent connections for each player.
*   **Concurrency:** Uses a **Fan-Out pattern**. Each client has a dedicated writer Goroutine and a buffered channel. This ensures that one laggy client does not block the main game loop for everyone else.
*   **State Management:** Shared game state (player locations) is protected by `sync.Mutex`.
*   **Protocol:** Custom headers (`MAP:` and `MSG:`) allow multiplexing the game grid and chat messages over a single TCP stream.
*   **Proximity Chat:** Chat messages are only broadcast to players within a specific radius on the grid, rather than globally.

## Project Structure
*   `cmd/server`: Entry point.
*   `internal/game`: Core logic (Connection handling, broadcasting, rendering).
*   `internal/domain`: Data structures (Player, Map constants).
*   `client`: A separate TUI application to connect to the server.

## Usage

1. **Start the Server**
   ```bash
   go run cmd/server/main.go
   ```

2. **Start a Client** (Run this in multiple terminals)
   ```bash
   go run client/main.go
   ```

3. **Controls**
    *   **Move:** `/w`, `/a`, `/s`, `/d`
    *   **Chat:** Type any text and hit Enter.
