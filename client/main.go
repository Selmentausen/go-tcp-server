package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, _ := net.Dial("tcp", "localhost:8888")
	defer conn.Close()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			conn.Write([]byte(text + "\n"))
		}
	}()

	serverReader := bufio.NewScanner(conn)
	for serverReader.Scan() {
		fmt.Println(">>", serverReader.Text())
	}
}
