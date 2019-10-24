// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Identity package will manage the Device Identity encoding / decoding
package identity

import (
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

type PeerIdentityJson struct {
	Name    string
	Id      string
	PrivKey string
	PubKey  string
}

func CreateIdentity(name string, privKey crypto.PrivKey) PeerIdentityJson {
	ID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		panic(err)
	}
	_ = ID

	pvtBytes, err := privKey.Raw()
	if err != nil {
		panic(err)
	}
	_ = pvtBytes
	pubBytes, err := privKey.GetPublic().Raw()
	if err != nil {
		panic(err)
	}
	_ = pubBytes

	return PeerIdentityJson{
		Name: name,
		Id: ID.Pretty(),
		PrivKey: b64.StdEncoding.EncodeToString(pvtBytes),
		PubKey: b64.StdEncoding.EncodeToString(pubBytes),
	}
}

// Create json file with identity information
func CreateIdentityJson(privKey crypto.PrivKey) string {
	identityJson := CreateIdentity("PeerVault device identity", privKey)
	idJson, err := json.MarshalIndent(identityJson, "", " ")
	if err != nil {
		panic(err)
	}

	return string(idJson)
}

func ReadIdentityJson(filePath string) (crypto.PrivKey, crypto.PubKey, error) {
	// Open our jsonFile
	jsonFile, err := os.Open(filePath)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, nil, err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var peerIdentity PeerIdentityJson

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &peerIdentity)

	privKeyByte, err := b64.StdEncoding.DecodeString(peerIdentity.PrivKey)
	if err != nil {
		return nil, nil, err
	}

	pvtKey, err := crypto.UnmarshalSecp256k1PrivateKey(privKeyByte)
	if err != nil {
		return nil, nil, err
	}

	return pvtKey, pvtKey.GetPublic(), nil
}
