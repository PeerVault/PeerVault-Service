// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Control package will manage all the communication with
// the local client of desktop, CLI or Graphic
package control

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Power-LAB/PeerVault/business/owner"
)

const (
	timeout = 10 * time.Second
)

func Listen(address* string) {
	fmt.Println("listen from control")

	// GET / POST owner information
	http.HandleFunc("/owner", owner.Controller)
	// POST owner SEED information
	http.HandleFunc("/owner/seed", owner.ControllerSeed)

	s := &http.Server{
		Addr:           *address,
		Handler:        nil,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
