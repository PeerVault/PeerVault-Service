// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Peer package will manage all the communication with
// the EXTERNAL Peers (libp2p) to exchange password with other Peer
// same of different owner
package peer

import (
	"fmt"
	"github.com/Power-LAB/PeerVault/crypto"
)

func Listen() {
	seed := &crypto.Seed{}
	seed.CreateMnemonic()
	fmt.Println("listen from peer", seed.Mnemonic)
}
