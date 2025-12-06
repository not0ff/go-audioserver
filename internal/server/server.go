package server

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/not0ff/go-audioserver/internal/audio"
	"github.com/not0ff/go-audioserver/pkg/message"
	"github.com/not0ff/go-audioserver/pkg/packet"
)

type Server interface {
	Cleanup()
	HandleClient(c net.Conn)
	ProcessMessage(m message.Message)
	Start(addr string)
}

type UdsServer struct {
	sockAddr string
	player   *audio.AudioPlayer
	quitChan chan os.Signal
}

func NewUdsServer(addr string) *UdsServer {
	return &UdsServer{sockAddr: addr, player: audio.NewAudioPlayer(), quitChan: make(chan os.Signal, 1)}
}

func (s *UdsServer) Start() {
	s.Cleanup()

	l, err := net.Listen("unix", s.sockAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer l.Close()
	log.Printf("Listening on %s", s.sockAddr)

	signal.Notify(s.quitChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-s.quitChan
		log.Print("Shutting down...")
		close(s.quitChan)
		s.Cleanup()
		os.Exit(1)
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		go s.HandleClient(conn)
	}
}

// Remove remaining socket file
func (s *UdsServer) Cleanup() {
	if _, err := os.Stat(s.sockAddr); !errors.Is(err, os.ErrNotExist) {
		if err := os.RemoveAll(s.sockAddr); err != nil {
			log.Fatal(err)
		}
	}
}

// Handle incoming connections
func (s *UdsServer) HandleClient(c net.Conn) {
	defer c.Close()
	log.Printf("Client connected [%s]", c.RemoteAddr().Network())

	for {
		p := packet.NewPacket()
		if err := p.Read(c); err != nil {
			if err != io.EOF {
				log.Printf("Error receiving packet: %s", err)
			}
			break
		}

		msg := message.NewMessage()
		if err := msg.Decode(p.Data); err != nil {
			log.Printf("Error decoding message: %s", err)
			break
		}

		go s.ProcessMessage(msg)
	}

	log.Print("Connection closed")
}

func (s *UdsServer) runAction(act int, p message.Payload) error {
	// It should be a valid type when passed
	switch act {
	case 0:
		return s.player.Play(p.(*message.PlayPayload))
	case 1:
		return s.player.Pause(p.(message.IdPayload).Id)
	case 2:
		return s.player.Resume(p.(message.IdPayload).Id)
	case 3:
		return s.player.Quit(p.(message.IdPayload).Id)
	default:
		return errors.New("invalid action id")
	}
}

// Executes funtion according to message action
func (s *UdsServer) ProcessMessage(m *message.Message) {
	if m.Action == 0 {
		pp := &message.PlayPayload{}
		if err := json.Unmarshal(m.Payload, pp); err != nil {
			log.Print("Cannot read invalid payload")
		}
		log.Printf("Received: [Action:%v; Payload:%s]", m.Action, pp)
		if err := s.runAction(m.Action, pp); err != nil {
			log.Printf("Error running action: %s", err)
		}
	} else if m.Action >= 1 && m.Action <= 3 {
		p := message.IdPayload{}
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			log.Print("Cannot read invalid payload")
		}
		log.Printf("Received: [Action:%v; Payload:%+v]", m.Action, p)
		if err := s.runAction(m.Action, p); err != nil {
			log.Printf("Error running action: %s", err)
		}
	}
}
