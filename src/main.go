package main

import (
  "github.com/Power-LAB/PeerVault/control"
  "github.com/Power-LAB/PeerVault/peer"
)

func main() {
  go control.Listen()
  go peer.Listen()

	select {}
}
