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
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Power-LAB/PeerVault/business/owner"
	"github.com/Power-LAB/PeerVault/business/secret"
	"github.com/Power-LAB/PeerVault/communication/event"
	"github.com/Power-LAB/PeerVault/crypto"
	"github.com/Power-LAB/PeerVault/database"
	"github.com/libp2p/go-libp2p-core/network"
	"go.etcd.io/bbolt"
	"strings"
)

type ShareRequest struct {
	Uuid string
	Sender string
	Receiver string
	Expiration string
	KeyPath string
	Approved bool
}

type Share struct {
	Uuid string
	Sender string
	Receiver string
	Expiration string
	KeyPath string
}

type ShareResponse struct {
	Uuid string
	Sender string
	Approved bool
}

type ShareResponseData struct {
	Uuid string
	Sender string
	Secret secret.Secret
}

type Error int

func (k Error) Error() (msg string) {
	switch k {
	case ErrorShareNotFound:
		msg = "The Secret key path, namespace and key name was not found"
	}
	return fmt.Sprintf("%s (%d)", msg, k)
}

const (
	ErrorShareNotFound = Error(1)
)

var (
	requests = make(map[string]ShareRequest)
)

func ApproveLocalRequest(uuid string) {
	if shareRequest, ok := requests[uuid]; ok {
		shareRequest.Approved = true
		requests[uuid] = shareRequest
	}
}

// Receive new request for sharing password
func secretShareRequestProtocol(s network.Stream) {
	log.Debug("Peer secretShareRequestProtocol")
	log.Debugf("remote are: %s\n", s.Conn().RemotePeer())
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	buf, err := rw.ReadString('\n')
	if err != nil {
		log.Error(err)
		return
	}
	shareRequest := &ShareRequest{}
	// Verify and decode share request data
	decoder := json.NewDecoder(strings.NewReader(buf))
	err = decoder.Decode(&shareRequest)
	if err != nil {
		log.Error(err)
		return
	}

	if shareRequest.Sender != s.Conn().RemotePeer().Pretty() {
		log.Error("Share request corrupted, sender and remote peer are different")
		return
	}
	requests[shareRequest.Uuid] = *shareRequest

	_ = event.Write(event.Message{
		Type: "secret.share.request",
		Data: map[string]string {
			"Sender": s.Conn().RemotePeer().Pretty(),
			"Uuid": shareRequest.Uuid,
			"SecretPath": shareRequest.KeyPath,
			"Expiration": shareRequest.Expiration,
		},
	})
	err = s.Close()
	if err != nil {
		log.Error(err)
	}
}

// Receive response confirmation for password sharing
func secretShareResponseProtocol(s network.Stream) {
	log.Debug("Peer secretShareResponseProtocol")
	log.Debugf("remote are: %s\n", s.Conn().RemotePeer())
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	buf, err := rw.ReadString('\n')
	if err != nil {
		log.Error(err)
		return
	}
	shareResponse := &ShareResponse{}
	// Verify and decode share request data
	decoder := json.NewDecoder(strings.NewReader(buf))
	err = decoder.Decode(&shareResponse)
	if err != nil {
		log.Error(err)
		return
	}
	_ = s.Close()

	share, err := getDbShareRequest([]byte(shareResponse.Uuid))
	if err != nil {
		if err == ErrorShareNotFound {
			log.Errorf("share request not found with uuid %s", shareResponse.Uuid)
		} else {
			log.Error(err)
		}
		return
	}
	if shareResponse.Approved == false {
		_ = event.Write(event.Message{
			Type: "secret.share.declined",
			Data: map[string]string {
				"Receiver": share.Receiver,
				"SecretPath": share.KeyPath,
				"Expiration": share.Expiration,
			},
		})
		return
	}
	secretData, err := secret.FetchSecret([]byte(share.KeyPath))
	o := owner.Owner{}
	if o.FetchOwner() != nil {
		log.Error(err)
		return
	}
	id, err := o.GetIdentity()
	if err != nil {
		log.Error(err)
		return
	}
	plainText, err := crypto.DecryptAes(id.GetChildKeyAsByte(), []byte(secretData.Value))
	secretData.Value = string(plainText)
	if err != nil {
		log.Error(err)
		return
	}
	responseData := &ShareResponseData{
		Uuid: share.Uuid,
		Sender: share.Sender,
		Secret: secretData,
	}
	secretJson, _ := json.Marshal(responseData)
	err = Dial(share.Receiver, PidShareSecret, secretJson)
	if err != nil {
		log.Error(err)
		return
	}
}

// Receive the password after share request exchange has been done
func secretProtocol(s network.Stream) {
	log.Debug("Peer secretProtocol")
	log.Debugf("remote are: %s\n", s.Conn().RemotePeer())

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	buf, err := rw.ReadString('\n')
	if err != nil {
		log.Error(err)
		return
	}
	shareResponseData := &ShareResponseData{}
	// Verify and decode share request data
	decoder := json.NewDecoder(strings.NewReader(buf))
	err = decoder.Decode(&shareResponseData)
	if err != nil {
		log.Error(err)
		return
	}
	_ = s.Close()
	shareRequest, ok := requests[shareResponseData.Uuid]
	if !ok {
		log.Errorf("share request not found with uuid %s", shareResponseData.Uuid)
	}
	if shareRequest.Approved == false {
		log.Error("This should not append, share request has not been approved")
		return
	}
	if shareRequest.Approved == true {
		o := owner.Owner{}
		if o.FetchOwner() != nil {
			log.Error(err)
			return
		}
		id, err := o.GetIdentity()
		if err != nil {
			log.Error(err)
			return
		}
		cipherSecretValue, err := crypto.EncryptAes(id.GetChildKeyAsByte(), []byte(shareResponseData.Secret.Value))
		shareResponseData.Secret.Value = string(cipherSecretValue)
		log.Debug(shareResponseData.Secret)
		err = shareResponseData.Secret.CreateSecret()
		if err != nil {
			log.Error(err)
			return
		}
		_ = event.Write(event.Message{
			Type: "secret.share.created",
			Data: map[string]string {
				"Sender": shareRequest.Sender,
				"SecretPath": shareRequest.KeyPath,
				"Type": string(shareResponseData.Secret.Type),
				"Description": shareResponseData.Secret.Description,
			},
		})
	}
}

func getDbShareRequest(uuid []byte) (Share, error) {
	share := &Share{}
	db, err := database.GetConnection()
	if err != nil {
		return *share, err
	}

	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("share"))
		if b == nil {
			return ErrorShareNotFound
		}
		buf := b.Get(uuid)
		if buf == nil {
			return ErrorShareNotFound
		}
		return json.Unmarshal(buf, share)
	})
	if err != nil {
		return *share, err
	}

	return *share, nil
}