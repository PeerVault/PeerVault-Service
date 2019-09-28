// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Peer package will manage all the communication with
// the EXTERNAL Peers (libp2p) to exchange password with other Peer
// same of different owner
package peer

import (
	"github.com/Power-LAB/PeerVault/crypto"
	"log"
)

func Listen() {
	seed := &crypto.Seed{}
	seed.CreateMnemonic()
	log.Printf("listen from peer %s", seed.Mnemonic)
}
