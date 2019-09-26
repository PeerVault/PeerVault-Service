// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Control package will manage all the communication with
// the local client of desktop, CLI or Graphic
package control

import (
  "fmt"
	"flag"
	"log"
	"net/http"

  "github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func Listen() {
	fmt.Println("listen from control")
	flag.Parse()
  log.SetFlags(0)
  http.HandleFunc("/echo", echo)
  log.Fatal(http.ListenAndServe(*addr, nil))

}
