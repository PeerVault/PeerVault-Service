package main

import (
	"flag"
	"github.com/Power-LAB/PeerVault/communication/control"
	"github.com/Power-LAB/PeerVault/communication/event"
	"github.com/Power-LAB/PeerVault/communication/peer"
	"log"
	"os"
)

func main() {
	apiAddress := flag.String("apiAddr", ":4444", "http api service address")
	wsAddress := flag.String("wsAddr", "localhost:5555", "WebSocket event service address")
	relayHost := flag.String("relay", "", "Relay Host URL")
	flag.Parse()

	if *relayHost == "" {
		log.Fatal("Please provide relay host with --relay option")
		os.Exit(0)
	}

	// Start websocket events
	go event.Listen(wsAddress)

	// Start API
	go control.Listen(apiAddress)

	// Start peer
	go peer.Listen(*relayHost)

	select {}
}
