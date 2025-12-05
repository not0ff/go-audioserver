package main

import (
	"flag"
	"log"
	"net"

	"github.com/not0ff/go-audioserver/pkg/message"
	"github.com/not0ff/go-audioserver/pkg/packet"
)

// ! NOTE: That's a simple client meant only for testing some basic functions

var (
	sockAddr    = flag.String("sockFile", "/tmp/go-audioserver.sock", "Path to socket address")
	audioFile   = flag.String("audioFile", "", "Path to audio file to send")
	audioFormat = flag.String("audioFormat", "", "Format of provided audio file")
)

func main() {
	flag.Parse()
	if *audioFile == "" || *audioFormat == "" {
		log.Fatal("Missing arguments. Check --help")
	}

	conn, err := net.Dial("unix", *sockAddr)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer conn.Close()

	msg := message.NewMessage()
	msg.Action = 0
	args := message.PlayPayload{
		Id:     0,
		Format: *audioFormat,
		Path:   *audioFile,
	}
	// args := message.IdPayload{Id: 0}
	if err := msg.SetPayload(args); err != nil {
		log.Fatal(err)
	}

	d, err := msg.Encode()
	if err != nil {
		log.Fatal(err)
	}

	p := packet.NewPacket()
	p.SetData(d)
	if err := p.Write(conn); err != nil {
		log.Printf("Error sending packet: %s", err)
		return
	}
}
