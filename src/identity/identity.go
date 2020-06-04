// Copyright (c) 2020, Pierre Tomasina
// Use of this source code is governed by a GNU AGPLv3
// license that can be found in the LICENSE file.
//
// Identity package will manage the Device Identity encoding / decoding
package identity

import (
	b64 "encoding/base64"
	"encoding/json"
	"github.com/PeerVault/PeerVault-Service/crypto"
	"github.com/op/go-logging"

	p2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

var (
	log = logging.MustGetLogger("peerVaultLogger")
)

type PeerIdentity struct {
	Name     string
	Id       string
	ChildKey string // Key represents a bip32 extended key 33bytes into string
	PrivKey  string
	PubKey   string
}

func GetIdentity(QmPeerId string) (PeerIdentity, error) {
	peerIdentity := &PeerIdentity{}
	keychain := crypto.Keychain{}
	err := keychain.CreateOrOpen()
	if err != nil {
		return *peerIdentity, err
	}

	idJson, err := keychain.Get(QmPeerId, "Owner")
	if err != nil {
		log.Debugf("The private key of QmPeerId %s has not been found on KeyChain", QmPeerId)
		return *peerIdentity, err
	}

	err = json.Unmarshal(idJson, peerIdentity)
	return *peerIdentity, err
}

func (p *PeerIdentity) GetChildKeyAsByte() []byte {
	childKey, _ := b64.StdEncoding.DecodeString(p.ChildKey)
	return  childKey
}


func (p PeerIdentity) GetCryptoPrivateKey() (p2pCrypto.PrivKey, error) {
	privKeyByte, err := b64.StdEncoding.DecodeString(p.PrivKey)
	if err != nil {
		return nil, err
	}
	pvtKey, err := p2pCrypto.UnmarshalSecp256k1PrivateKey(privKeyByte)
	if err != nil {
		return nil, err
	}
	return pvtKey, nil
}

func CreateIdentity(name string, privKey p2pCrypto.PrivKey, childKey []byte) (PeerIdentity, error) {
	ID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return PeerIdentity{}, err
	}

	pvtBytes, err := privKey.Raw()
	if err != nil {
		return PeerIdentity{}, err
	}

	pubBytes, err := privKey.GetPublic().Raw()
	if err != nil {
		return PeerIdentity{}, err
	}

	return PeerIdentity{
		Name: name,
		Id: ID.Pretty(),
		ChildKey: b64.StdEncoding.EncodeToString(childKey),
		PrivKey: b64.StdEncoding.EncodeToString(pvtBytes),
		PubKey: b64.StdEncoding.EncodeToString(pubBytes),
	}, nil
}

func (p PeerIdentity) SaveIdentity() error {
	idJson, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		return err
	}

	keychain := crypto.Keychain{}
	err = keychain.CreateOrOpen()
	if err != nil {
		return err
	}
	return keychain.Put(p.Id, idJson, "Owner", false)
}

func DeleteIdentity(QmPeerId string) error {
	keychain := crypto.Keychain{}
	err := keychain.CreateOrOpen()
	if err != nil {
		return err
	}
	return keychain.Delete(QmPeerId)
}