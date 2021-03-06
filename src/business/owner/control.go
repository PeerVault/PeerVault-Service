// Package owner will manage owner information
//
// Copyright (c) 2020, Pierre Tomasina
// Use of this source code is governed by a GNU AGPLv3
// license that can be found in the LICENSE file.
package owner

import (
	"encoding/json"
	"fmt"
	"github.com/PeerVault/PeerVault-Service/crypto"
	"github.com/PeerVault/PeerVault-Service/identity"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
)

var (
	log = logging.MustGetLogger("peerVaultLogger")
)

// Manage Owner GET / POST
func Controller(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
		case http.MethodGet:
			getOwner(w, r)
		case http.MethodPatch:
			if !PasswordVerification(r, false) {
				http.Error(w, "{\"error\": \"X-OWNER-CODE is required\"}", http.StatusUnauthorized)
				return
			}
			updateOwner(w, r)
		case http.MethodPost:
			createOwner(w, r)
		case http.MethodDelete:
			requestDeleteOwner(w, r)
		default:
			http.Error(w, "Invalid request method.", 405)
	}
}

//  Manage seed restoration
func ControllerSeed(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		restoreOwner(w, r)
	default:
		http.Error(w, "Invalid request method.", 405)
	}
}

// Retrieved owner information
func getOwner(w http.ResponseWriter, r *http.Request) {
	exist, err := IsOwnerExist()
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	if exist == false {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusFailedDependency)
		_, _ = w.Write([]byte("{\"error\": \"Owner not existing, you must create one first\"}"))
		return
	}

	o := &Owner{}
	err = o.FetchOwner()
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	buf, err := json.Marshal(o)
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(buf)
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
}

// Create new owner
func createOwner(w http.ResponseWriter, r *http.Request) {
	// Verify and decode Owner input data
	decoder := json.NewDecoder(r.Body)
	var o Owner
	err := decoder.Decode(&o)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("{\"error\": \"Payload must be struct Owner\"}"))
		return
	}
	if o.AskPassword > 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("{\"error\": \"AskPassword must be 0,1,2\"}"))
	}

	exist, err := IsOwnerExist()
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	if exist == true {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusFailedDependency)
		_, _ = w.Write([]byte("{\"error\": \"Owner already exist, you should update it with patch\"}"))
		return
	}

	seed := &crypto.Seed{}
	seed.CreateSeed()
	master, err := seed.CreateMasterKey()
	if err != nil {
		log.Debug("Master key fail to create")
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	child, err := crypto.CreateChildKey(master)
	if err != nil {
		log.Debug("Child key fail to create")
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pvtKey, err := crypto.BipKeyToLibp2p(child)
	if err != nil {
		log.Debug("BipKey to Libp2p convertion error")
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	peerIdentity, err := identity.CreateIdentity(o.DeviceName, pvtKey, child.Key)
	if err != nil {
		log.Debug("Error during identity creation")
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	o.QmPeerId = peerIdentity.Id

	// Save owner in DB
	err = o.PutOwner()
	if err != nil {
		log.Debug("PutOwner error")
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	err = peerIdentity.SaveIdentity()
	if err != nil {
		log.Debug("SaveIdentity error")
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("{\"seed\": \"%s\"}", seed.Mnemonic)))
}

// Change owner information
func updateOwner(w http.ResponseWriter, r *http.Request) {
	o := &Owner{}
	err := o.FetchOwner()
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	// Verify and decode Owner input data
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&o)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("{\"error\": \"Payload must be struct Owner\"}"))
		return
	}

	// Save owner in DB
	err = o.PutOwner()
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// Create owner from SEED
func restoreOwner(w http.ResponseWriter, r *http.Request) {
	exist, err := IsOwnerExist()
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	if exist == true {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "{\"error\": \"Owner already exist, you must delete it to restore new seed\"}", http.StatusFailedDependency)
		return
	}

	// Read body
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Verify and decode Owner input data
	var o Owner
	err = json.Unmarshal(body, &o)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "{\"error\": \"Payload must be struct Owner\"}", http.StatusBadRequest)
		return
	}

	m := struct {
		Mnemonic string
	}{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "{\"error\": \"Seed must be present to add existing account\"}", http.StatusBadRequest)
		return
	}

	// Restore seed
	seed := &crypto.Seed {
		Mnemonic: m.Mnemonic,
	}
	seed.CreateSeed()
	master, err := seed.CreateMasterKey()
	if err != nil {
		log.Debug("Master key fail to create")
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	child, err := crypto.CreateChildKey(master)
	if err != nil {
		log.Debug("Child key fail to create")
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pvtKey, err := crypto.BipKeyToLibp2p(child)
	if err != nil {
		log.Debug("BipKey to Libp2p convertion error")
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	peerIdentity, err := identity.CreateIdentity(o.DeviceName, pvtKey, child.Key)
	if err != nil {
		log.Debug("Error during identity restore creation")
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	o.QmPeerId = peerIdentity.Id

	// Save owner in DB
	err = o.PutOwner()
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	err = peerIdentity.SaveIdentity()
	if err != nil {
		log.Debug("Error during restore identity save")
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Generate a code to confirm deletion of owner and therefore any information of the current device
func requestDeleteOwner(w http.ResponseWriter, r *http.Request) {
	// TODO must notify other peer of the deletion of device
	http.Error(w, "Delete not yet implemented", http.StatusServiceUnavailable)
}