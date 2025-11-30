package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net"
	"os"

	"github.com/not0ff/go-audioserver/internal/message"
	"github.com/not0ff/go-audioserver/internal/packet"
)

const SockAddr = "/tmp/uds.sock"
const DataPath = "/Audio/audio.mp3"

// ! NOTE: That's a simple client meant only for testing some basic functions

func main() {
	conn, err := net.Dial("unix", SockAddr)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer conn.Close()

	audioId := 1102

	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	defer w.Close()

	data, err := os.Open(DataPath)
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(w, data)

	msg := message.NewMessage()
	msg.Action = 0
	args := message.PlayPayload{
		Id:     audioId,
		Format: "mp3",
		Data:   b.Bytes(),
		Volume: -2,
	}
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
	// log.Printf("Sent %s", p)

	// time.Sleep(time.Second * 5)
	// msg = message.NewMessage()
	// msg.Action = 1
	// args2 := message.IdPayload{Id: audioId}
	// if err := msg.SetPayload(args2); err != nil {
	// 	log.Fatal(err)
	// }

	// d, err = msg.Encode()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// p = packet.NewPacket()
	// p.SetData(d)
	// if err := p.Write(conn); err != nil {
	// 	log.Printf("Error sending packet: %s", err)
	// 	return
	// }
	// log.Printf("Sent %s", p)
}
