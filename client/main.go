package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app       *tview.Application
	mapView   *tview.TextView
	chatView  *tview.TextView
	inputView *tview.InputField
	conn      net.Conn
)

func main() {
	var err error
	conn, err = net.Dial("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	app = tview.NewApplication()

	// Map Area
	mapView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("Waiting for map...")
	mapView.SetBorder(true).SetTitle(" Map ")

	// Chat Area
	chatView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	chatView.SetBorder(true).SetTitle(" Chat ")

	// Input Area
	inputView = tview.NewInputField().
		SetLabel("Command: ").
		SetFieldWidth(0).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				text := inputView.GetText()
				if text != "" {
					conn.Write([]byte(text + "\n"))
					inputView.SetText("")
				}
			}
		})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mapView, 0, 3, false).
		AddItem(chatView, 0, 2, false).
		AddItem(inputView, 3, 1, true)

	go listenToServer()

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

func listenToServer() {
	buffer := make([]byte, 4096)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		data := string(buffer[:n])

		if strings.Contains(data, "MAP:") {
			parts := strings.Split(data, "MAP:")
			if len(parts) > 1 {
				mapData := parts[len(parts)-1]
				if idx := strings.Index(mapData, "MSG:"); idx != -1 {
					mapData = mapData[:idx]
				}

				app.QueueUpdateDraw(func() {
					mapView.SetText(tview.TranslateANSI(mapData))
				})
			}
		}

		if strings.Contains(data, "MSG:") {
			parts := strings.Split(data, "MSG:")
			for _, p := range parts {
				if len(p) > 0 && !strings.Contains(p, ".") {
					app.QueueUpdateDraw(func() {
						fmt.Fprint(chatView, p)
						chatView.ScrollToEnd()
					})
				}
			}
		}
	}
}
