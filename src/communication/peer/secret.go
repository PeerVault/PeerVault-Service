// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Peer package will manage all the communication with
// the EXTERNAL Peers (libp2p) to exchange password with other Peer
// same of different owner
//
// Secret will focus on Protocol regarding secret exchange
package peer

import (
	"fmt"
	"github.com/Power-LAB/PeerVault/communication/event"
	"github.com/libp2p/go-libp2p-core/network"
)

func secretProtocol(s network.Stream) {
	fmt.Printf("Meow! It worked! remote are: %s\n", s.Conn().RemotePeer())
	err := event.Write(event.Message{
		Type: "message",
		Data: nil,
	})
	if err != nil {
		fmt.Printf("SecretProtocol ERROR: %s", err.Error())
	}
	s.Close()
}