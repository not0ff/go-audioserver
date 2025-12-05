package main

import (
	"flag"

	"github.com/not0ff/go-audioserver/internal/server"
)

var SockAddr = flag.String("sock", "/tmp/go-audioserver.sock", "Path to socket address")

func main() {
	flag.Parse()
	s := server.NewUdsServer(*SockAddr)
	s.Start()
}
