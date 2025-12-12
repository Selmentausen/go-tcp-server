package domain

const (
	MAP_WIDTH   = 20
	MAP_HEIGHT  = 10
	CHAT_RADIUS = 5
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

type Player struct {
	Name    string
	X       int
	Y       int
	Color   string
	MsgChan chan string
}

func NewPlayer(name string, x, y int, color string) *Player {
	return &Player{
		Name:    name,
		X:       x,
		Y:       y,
		Color:   color,
		MsgChan: make(chan string, 10),
	}
}
