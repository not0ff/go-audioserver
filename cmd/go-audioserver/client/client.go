package main

import (
	"log"
	"net"

	"github.com/not0ff/go-audioserver/internal/message"
)

const SockAddr = "/tmp/uds.sock"

func main() {
	conn, err := net.Dial("unix", SockAddr)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer conn.Close()

	s := "Hello from the void"
	msg := message.NewMessage()
	msg.SetData([]byte(s))
	if err := msg.Write(conn); err != nil {
		log.Printf("Error sending message: %s", err)
		return
	}
	log.Printf("Sent %s", msg)

	if err := msg.Read(conn); err != nil {
		log.Printf("Error receiving message: %s", err)
		return
	}
	log.Printf("Received %s", msg)
}
