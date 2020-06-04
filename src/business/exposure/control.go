// Copyright (c) 2020, Pierre Tomasina
// Use of this source code is governed by a GNU AGPLv3
// license that can be found in the LICENSE file.
//
// Exposure package will manage the secret exposure to the client
package exposure

import (
	"encoding/json"
	"fmt"
	"github.com/Power-LAB/PeerVault/business/owner"
	"github.com/Power-LAB/PeerVault/business/secret"
	"github.com/Power-LAB/PeerVault/communication/peer"
	"github.com/Power-LAB/PeerVault/crypto"
	"github.com/google/uuid"
	"github.com/op/go-logging"
	"net/http"
	"path"
	"time"
)

var (
	log = logging.MustGetLogger("peerVaultLogger")
)

// Manage Secret GET / POST
func Controller(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !owner.PasswordVerification(r, true) {
		http.Error(w, "{\"error\": \"X-OWNER-CODE is required\"}", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getSecretValue(w, r)
	default:
		http.Error(w, "Invalid request method.", 405)
	}
}

// Manage Exposure Request
// POST : Create a request for sharing secret with other peer
// GET : List requests, both sent and received
// DELETE : Decline or remove request
func ControllerRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodPost:
		createShareRequest(w, r)
	case http.MethodGet:
		getShareRequest(w, r)
	case http.MethodDelete:
		deleteShareRequest(w, r)
	case http.MethodPut:
		shareResponse(w, r)
	default:
		http.Error(w, "Invalid request method.", 405)
	}
}

func getSecretValue(w http.ResponseWriter, r *http.Request) {
	keyPath := []byte(path.Base(r.RequestURI))
	s, err := secret.FetchSecret(keyPath)

	if err == secret.ErrorSecretNotFound {
		http.Error(w, "{\"error\": \"Secret Not Found\"}", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error(err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	o := owner.Owner{}
	if o.FetchOwner() != nil {
		log.Notice(err)
		http.Error(w, "{\"error\": \"Owner not found\"}", http.StatusNotFound)
	}
	identity, err := o.GetIdentity()
	if err != nil {
		log.Debug("Cannot find Identity of current owner")
		log.Error(err)
		http.Error(w, "{\"error\": \"Cannot find Identity of current owner\"}", http.StatusInternalServerError)
	}
	plainText, err := crypto.DecryptAes(identity.GetChildKeyAsByte(), []byte(s.Value))
	if err != nil {
		log.Debug("Error during secret decipher")
		log.Error(err)
		http.Error(w, "{\"error\": \"Secret cannot be decrypted\"}", http.StatusInternalServerError)
	}
	s.Value = string(plainText)

	resultJson, _ := json.Marshal(s)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resultJson)
}

func createShareRequest(w http.ResponseWriter, r *http.Request) {
	shareRequest := &ShareRequest{}
	// Verify and decode Owner input data
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&shareRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("{\"error\": \"Payload must be struct of ShareRequest\"}"))
		return
	}

	// Check if KeyPath exist
	_, err = secret.FetchSecret([]byte(shareRequest.KeyPath))
	if err == secret.ErrorSecretNotFound {
		http.Error(w, "{\"error\": \"Not secret found with KeyPath specified\"}", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	// Get Owner
	o := owner.Owner{}
	if o.FetchOwner() != nil {
		log.Notice(err)
		http.Error(w, "{\"error\": \"Owner not found\"}", http.StatusNotFound)
	}

	// Create Share
	share := &Share{
		Uuid:       uuid.New().String(),
		Sender:     o.QmPeerId,
		Receiver:   shareRequest.Receiver,
		Expiration: time.Now().UTC().Add(shareRequest.ExpirationDelay * time.Hour).Format(time.RFC3339),
		KeyPath:    shareRequest.KeyPath,
	}
	err = share.Save()
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	resultJson, _ := json.Marshal(share)

	go func() {
		// Dial to receiver
		err := peer.Dial(shareRequest.Receiver, peer.PidShareRequest, resultJson)
		if err != nil {
			log.Error("Error during share request dial")
		}
	}()

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resultJson)
}

// Retrieved share request
func getShareRequest(w http.ResponseWriter, r *http.Request) {
	shares, err := FetchShares()
	if err != nil {
		fmt.Printf("INTERNAL ERROR: %s", err.Error())
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	resultJson, _ := json.Marshal(shares)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resultJson)
}

func deleteShareRequest(w http.ResponseWriter, r *http.Request) {
	share := &Share{
		Uuid: path.Base(r.RequestURI),
	}
	err := share.Delete()
	if err != nil {
		log.Debug("Error during share request deletion")
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func shareResponse(w http.ResponseWriter, r *http.Request) {
	shareResponse := &ShareResponse{}
	// Verify and decode Owner input data
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&shareResponse)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("{\"error\": \"Payload must be struct of ShareRequest\"}"))
		return
	}
	shareResponseJson, _ := json.Marshal(shareResponse)

	go func() {
		// Approve the request locally to trust data when secret will arrive
		if shareResponse.Approved {
			peer.ApproveLocalRequest(shareResponse.Uuid)
		}
		// Dial to sender to confirm share
		err := peer.Dial(shareResponse.Sender, peer.PidShareResponse, shareResponseJson)
		if err != nil {
			log.Error("Error during share response dial")
		}
	}()
	w.WriteHeader(http.StatusOK)
}