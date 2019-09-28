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

	"github.com/Power-LAB/PeerVault/event"
)

const (
	timeout = 10 * time.Second
)

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

func Listen(address* string) {
	fmt.Println("listen from control")

	http.HandleFunc("/owner", getOwner)

	s := &http.Server{
		Addr:           *address,
		Handler:        nil,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
