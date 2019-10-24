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
	"context"

	"github.com/Power-LAB/PeerVault/business/owner"
	"github.com/Power-LAB/PeerVault/communication/event"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

func Listen(relayHost string) {
	exist, err := owner.IsOwnerExist()
	if err != nil {
		fmt.Printf("PEER INTERNAL ERROR: %s", err.Error())
		return
	}
	if exist == false {
		fmt.Printf("Owner does not exist, Peer cannot start")
		return
	}
	o := &owner.Owner{}
	err = o.FetchOwner()
	if err != nil {
		fmt.Printf("PEER INTERNAL ERROR: %s", err.Error())
		return
	}

	// The context governs the lifetime of the libp2p node
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pvt, err := o.GetIdentity().GetCryptoPrivateKey()
	if err != nil {
		panic(err)
	}

	// Zero out the listen addresses for the host, so it can only communicate
	// via p2p-circuit for our example
	node, err := libp2p.New(
		ctx,
		libp2p.Identity(pvt),
		libp2p.ListenAddrs(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		panic(err)
	}

	// Creates relay peer.AddrInfo
	relayAddrInfo, err := p2pAddrInfo(relayHost)
	if err != nil {
		panic(err)
	}

	if err := node.Connect(context.Background(), *relayAddrInfo); err != nil {
		panic(err)
	}

	// Now, to test things, let's set up a protocol handler on node
	node.SetStreamHandler("/cats", func(s network.Stream) {
		fmt.Printf("Meow! It worked! remote are: %s\n", s.Conn().RemotePeer())
		err = event.Write(event.Message{
			Type: "message",
			Data: nil,
		})
		s.Close()
	})

	fmt.Println("Device QmID: ", node.ID().Pretty());
	// TODO SAVE QMID

	for _, addr := range node.Addrs() {
		fmt.Printf("Device Relay Addr: %s\n", addr.String())
	}

	select {}
}

// create peer addr info
func p2pAddrInfo(addrStr string) (*peer.AddrInfo, error) {
	addr, err := ma.NewMultiaddr(addrStr)
	if err != nil {
		panic(err)
	}
	return peer.AddrInfoFromP2pAddr(addr)
}
