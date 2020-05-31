package secret

import (
	"encoding/json"
	"fmt"
	"github.com/Power-LAB/PeerVault/business/owner"
	"github.com/Power-LAB/PeerVault/crypto"
	"io/ioutil"
	"net/http"
	"path"
)

// Manage Secret GET / POST
func Controller(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !owner.PasswordVerification(r, false) {
		http.Error(w, "{\"error\": \"X-OWNER-CODE is required\"}", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getSecrets(w, r)
	case http.MethodPost:
		createSecret(w, r)
	case http.MethodDelete:
		deleteSecret(w, r)
	default:
		http.Error(w, "Invalid request method.", 405)
	}
}

// Retrieved secret information
func getSecrets(w http.ResponseWriter, r *http.Request) {
	secrets, err := FetchSecrets()
	if err != nil {
		fmt.Printf("INTERNAL ERROR: %s", err.Error())
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	resultJson, _ := json.Marshal(secrets)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resultJson)
}

func createSecret(w http.ResponseWriter, r *http.Request) {
	// Read body
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Verify and decode Secret input data
	var secret = &Secret{}
	err = json.Unmarshal(body, &secret)
	if err != nil {
		log.Error(err)
		http.Error(w, "{\"error\": \"Payload must be struct Secret\"}", http.StatusBadRequest)
		return
	}
	if !secret.assertSecretStruct() {
		log.Warning(err)
		http.Error(w, "{\"error\": \"Secret Namespace and Key must be alphanum with dash and underscore only allowed\"}", http.StatusBadRequest)
		return
	}
	o := owner.Owner{}
	if o.FetchOwner() != nil {
		log.Notice(err)
		http.Error(w, "{\"error\": \"Owner not found\"}", http.StatusNotFound)
	}

	// Encrypt the secret before saving into bbolt
	identity, err := o.GetIdentity()
	if err != nil {
		log.Debug("Cannot find Identity of current owner")
		log.Error(err)
		http.Error(w, "{\"error\": \"Cannot find Identity of current owner\"}", http.StatusInternalServerError)
	}
	cipherSecretValue, err := crypto.EncryptAes(identity.GetChildKeyAsByte(), []byte(secret.Value))
	if err != nil {
		log.Debug("Error during secret encryption")
		log.Error(err)
		http.Error(w, "{\"error\": \"Secret cannot be encrypted\"}", http.StatusInternalServerError)
	}
	secret.Value = string(cipherSecretValue)

	err = secret.CreateSecret()
	if err != nil {
		log.Debug("Cannot create secret in local database")
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func deleteSecret(w http.ResponseWriter, r *http.Request) {
	keyPath := path.Base(r.RequestURI)
	err := DeleteSecret(keyPath)
	if err != nil {
		log.Debug("Error during delete of Secret in local database")
		log.Error(err)
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}