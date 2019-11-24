// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Control package will manage all the communication with
// the local client of desktop, CLI or Graphic
package control

import (
	"log"
	"net/http"
	"time"

	"github.com/Power-LAB/PeerVault/business/owner"
	"github.com/Power-LAB/PeerVault/business/secret"
)

const (
	timeout = 10 * time.Second
)

func Listen(address* string) {
	// GET / POST owner information
	http.HandleFunc("/owner", owner.Controller)
	// POST owner SEED information
	http.HandleFunc("/owner/seed", owner.ControllerSeed)
	// GET / POST secret information
	http.HandleFunc("/secret", secret.Controller)
	http.HandleFunc("/secret/", secret.Controller)

	s := &http.Server{
		Addr:           *address,
		Handler:        nil,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
