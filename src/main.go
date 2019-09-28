package main

import (
	"flag"
	"github.com/Power-LAB/PeerVault/control"
	"github.com/Power-LAB/PeerVault/event"
	"github.com/Power-LAB/PeerVault/peer"
)

func main() {
	apiAddress := flag.String("apiAddr", ":4444", "http api service address")
	wsAddress := flag.String("wsAddr", "localhost:5555", "WebSocket event service address")
	flag.Parse()

	// Start websocket events
	go event.Listen(wsAddress)

	// Start API
	go control.Listen(apiAddress)

	// Start peer
	// TODO must be auto started only if config are present
	go peer.Listen()

	select {}
}
