package main

import (
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/not0ff/go-audioserver/internal/message"
)

const SockAddr = "/tmp/uds.sock"

func main() {
	cleanup()

	l, err := net.Listen("unix", SockAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer l.Close()
	log.Printf("Listening on %s", SockAddr)

	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-q
		log.Print("Shutting down...")
		close(q)
		cleanup()
		os.Exit(1)
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		go handleConn(conn)
	}
}

func cleanup() {
	if _, err := os.Stat(SockAddr); !errors.Is(err, os.ErrNotExist) {
		if err := os.RemoveAll(SockAddr); err != nil {
			log.Fatal(err)
		}
	}
}

func handleConn(c net.Conn) {
	defer c.Close()
	log.Printf("Client connected [%s]", c.RemoteAddr().Network())

	msg := message.NewMessage()
	if err := msg.Read(c); err != nil {
		log.Printf("Error receiving message: %s", err)
		return
	}
	log.Printf("Received %s", msg)

	if err := msg.Write(c); err != nil {
		log.Printf("Error sending message: %s", err)
		return
	}
	log.Printf("Sent %s", msg)

	log.Print("Connection closed")
}
