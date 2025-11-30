package main

import (
	"github.com/not0ff/go-audioserver/internal/server"
)

const SockAddr = "/tmp/uds.sock"

func main() {
	s := server.NewUdsServer(SockAddr)
	s.Start()
}
