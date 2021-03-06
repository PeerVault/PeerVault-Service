// Copyright (c) 2020, Pierre Tomasina
// Use of this source code is governed by a GNU AGPLv3
// license that can be found in the LICENSE file.
//
// Package control will manage all the communication with the client (GUI / CLI)
package control

import (
	"net/http"
	"time"
	"github.com/PeerVault/PeerVault-Service/business/exposure"
	"github.com/PeerVault/PeerVault-Service/business/owner"
	"github.com/PeerVault/PeerVault-Service/business/secret"
	"github.com/op/go-logging"
)

const (
	timeout = 10 * time.Second
)

var (
	log = logging.MustGetLogger("peerVaultLogger")
)

func Listen(address* string) {
	log.Info("listen from control")

	// GET / POST owner information
	http.HandleFunc("/owner", owner.Controller)
	// POST owner SEED information
	http.HandleFunc("/owner/seed", owner.ControllerSeed)
	// GET / POST secret information
	http.HandleFunc("/secret", secret.Controller)
	http.HandleFunc("/secret/", secret.Controller)

	http.HandleFunc("/expose/", exposure.Controller)
	http.HandleFunc("/expose/request", exposure.ControllerRequest)
	http.HandleFunc("/expose/request/", exposure.ControllerRequest)

	s := &http.Server{
		Addr:           *address,
		Handler:        nil,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
