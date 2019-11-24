package secret

import (
	"encoding/json"
	"fmt"
	"github.com/Power-LAB/PeerVault/business/owner"
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
		fmt.Printf("INTERNAL ERROR: %s", err.Error())
		http.Error(w, "{\"error\": \"Payload must be struct Secret\"}", http.StatusBadRequest)
		return
	}
	err = secret.CreateSecret()
	if err != nil {
		fmt.Printf("INTERNAL ERROR: %s", err.Error())
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func deleteSecret(w http.ResponseWriter, r *http.Request) {
	key := path.Base(r.RequestURI)
	err := DeleteSecret(key)
	if err != nil {
		fmt.Printf("INTERNAL ERROR: %s", err.Error())
		http.Error(w, "{\"error\": \"internal server error\"}", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}