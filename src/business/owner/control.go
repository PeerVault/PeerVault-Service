package owner

import (
	"fmt"
	"net/http"

	"github.com/Power-LAB/PeerVault/communication/event"
)

// Manage Owner GET / POST
func Controller(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getOwner(w, r)
	case http.MethodPost:
		updateOwner(w, r)
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
	err := event.Write(event.Message{
		Type: "owner",
		Data: nil,
	})
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "Hello, %q", r.URL.Path)
}

// Change owner information
func updateOwner(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

// Change owner information
func restoreOwner(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}