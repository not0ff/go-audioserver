package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"flag"
	"log"
	"net"
	"os"

	"github.com/not0ff/go-audioserver/pkg/message"
	"github.com/not0ff/go-audioserver/pkg/packet"
)

var (
	sockAddr = flag.String("sockFile", "/tmp/go-audioserver.sock", "Path to socket address")
	action   = flag.Int("act", 0, "Specified id for action")
	id       = flag.Int("id", 1, "Id for new or existing playback")
	format   = flag.String("fmt", "", "Audio file format")
	path     = flag.String("path", "", "Path to audio file")
	asData   = flag.Bool("asData", false, "Send audio as bytes")
	volume   = flag.Int("vol", 0, "Modify audio volume")
	loop     = flag.Bool("loop", false, "Loop audio until stopped")
)

func writePlayPayload() (message.PlayPayload, error) {
	var p message.PlayPayload
	p.Id = *id
	if len(*format) == 0 {
		return p, errors.New("audio format not specified")
	}
	p.Format = *format

	if len(*path) == 0 {
		return p, errors.New("missing path to file")
	}
	if *asData {
		f, err := os.ReadFile(*path)
		if err != nil {
			return p, err
		}
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		defer w.Close()
		if _, err := w.Write(f); err != nil {
			return p, err
		}
		p.Data = buf.Bytes()
	} else {
		p.Path = *path
	}
	p.Volume = *volume
	p.Loop = *loop

	return p, nil
}

func main() {
	flag.Parse()

	conn, err := net.Dial("unix", *sockAddr)
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}
	defer conn.Close()

	msg := message.NewMessage()
	msg.Action = *action
	var p message.Payload
	if *action == 0 {
		p, err = writePlayPayload()
		if err != nil {
			log.Fatalf("Error writing payload: %s", err)
		}
	} else {
		p = message.IdPayload{Id: *id}
	}

	if err := msg.SetPayload(p); err != nil {
		log.Fatalf("Error setting payload: %s", err)
	}

	d, err := msg.Encode()
	if err != nil {
		log.Fatalf("Error decoding payload: %s", err)
	}

	pack := packet.NewPacket()
	pack.SetData(d)

	if err := pack.Write(conn); err != nil {
		log.Fatalf("Error sending packet: %s", err)
	}
}
