// Copyright (c) 2020, Pierre Tomasina
// Use of this source code is governed by a GNU AGPLv3
// license that can be found in the LICENSE file.
//
// Peer package will manage all the communication with
// the EXTERNAL Peers (libp2p) to exchange password with other Peer
// same or different owner
package peer

import (
	"bufio"
	"context"
	"fmt"
	"github.com/PeerVault/PeerVault-Service/identity"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/op/go-logging"

	"github.com/PeerVault/PeerVault-Service/business/owner"
	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	PidShareRequest protocol.ID = "/secret/share/request"
	PidShareResponse protocol.ID = "/secret/share/response"
	PidShareSecret protocol.ID = "/secret/share"
)

var (
	log = logging.MustGetLogger("peerVaultLogger")
	relayHost string
	node host.Host
)

func SetRelayHost(relay string) {
	relayHost = relay
}

func Listen() {
	if exist, _ := owner.IsOwnerExist(); exist == false {
		log.Warning("OWNER NOT SETUP, PEER P2P CANT CONNECT")
		return
	}

	log.Debug("Listen")
	peerIdentity, err := getPeerIdentity()
	if err != nil {
		log.Fatal(err)
	}

	// The context governs the lifetime of the libp2p node
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pvt, err := peerIdentity.GetCryptoPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	// Zero out the listen addresses for the host, so it can only communicate via p2p-circuit
	node, err = libp2p.New(
		ctx,
		libp2p.Identity(pvt),
		libp2p.ListenAddrs(),
		libp2p.EnableRelay(circuit.OptDiscovery),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Creates relay peer.AddrInfo
	relayAddrInfo, err := p2pAddrInfo(relayHost)
	if err != nil {
		log.Fatal(err)
	}

	if err := node.Connect(context.Background(), *relayAddrInfo); err != nil {
		log.Fatal(err)
	}

	// Define handle for sharing request protocol
	node.SetStreamHandler(PidShareRequest, secretShareRequestProtocol)
	node.SetStreamHandler(PidShareResponse, secretShareResponseProtocol)
	node.SetStreamHandler(PidShareSecret, secretProtocol)

	log.Info("listen from peer")
	log.Info(node.ID().Pretty())
	log.Info(node.Addrs())

	select {}
}

func Dial(recipient string, pid protocol.ID, data []byte) error {
	recipientPeerId, err := peer.IDB58Decode(recipient)
	if err != nil {
		return err
	}
	log.Debugf("recipientPeerId %s", relayHost + "/p2p-circuit/p2p/" + recipientPeerId.Pretty())
	ma.SwapToP2pMultiaddrs()
	relayAddr, err := ma.NewMultiaddr(relayHost + "/p2p-circuit/p2p/" + recipientPeerId.Pretty())
	if err != nil {
		log.Fatal(err)
	}

	recipientRelayInfo := peer.AddrInfo{
		ID: recipientPeerId,
		Addrs: []ma.Multiaddr{relayAddr},
	}

	// Connect node to recipient
	if err := node.Connect(context.Background(), recipientRelayInfo); err != nil {
		log.Error("fail connect to recipient using relay")
		return err
	}

	// we're connected!
	stream, err := node.NewStream(context.Background(), recipientPeerId, pid)
	if err != nil {
		log.Fatal("Fail opening protocol with other peer", err)
		return err
	}
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	_, err = rw.WriteString(fmt.Sprintf("%s\n", data))
	if err != nil {
		return err
	}
	err = rw.Flush()

	return err
}

func getPeerIdentity() (identity.PeerIdentity, error) {
	emptyIdentity := identity.PeerIdentity{}
	exist, err := owner.IsOwnerExist()
	if err != nil {
		log.Error("PEER INTERNAL ERROR: %s", err.Error())
		return emptyIdentity, err
	}
	if exist == false {
		log.Error("Owner does not exist, Peer cannot start")
		return emptyIdentity, err
	}
	o := &owner.Owner{}
	err = o.FetchOwner()
	if err != nil {
		log.Error("PEER INTERNAL ERROR: %s", err.Error())
		return emptyIdentity, err
	}
	return o.GetIdentity()
}

// create peer addr info
func p2pAddrInfo(addrStr string) (*peer.AddrInfo, error) {
	addr, err := ma.NewMultiaddr(addrStr)
	if err != nil {
		panic(err)
	}
	return peer.AddrInfoFromP2pAddr(addr)
}
