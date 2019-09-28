// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Event package will manage all the event dispatching to
// the local client using websocket
package event

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string `json:"type"`
	Data interface{} `json:"data"`
}

var upgrader = websocket.Upgrader{} // use default options
var connections = make([]websocket.Conn,0)

// Write response to all websocket client connected
func Write(message Message) error {
	msgJson, err := json.MarshalIndent(message, "", " ")
	if err != nil {
		return err
	}
	for _, conn := range connections {
		if err := conn.WriteJSON(string(msgJson)); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

// Process a new websocket communication
func process(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	connections = append(connections, *conn)

	result := Message {
		Type: "process-ok",
		Data: nil,
	}
	resultJson, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		panic(err)
	}

	if err := conn.WriteJSON(string(resultJson)); err != nil {
		log.Println(err)
		return
	}

}

func Listen(address* string) {
	fmt.Println("listen from event")
	log.SetFlags(0)
	http.HandleFunc("/process", process)
	log.Fatal(http.ListenAndServe(*address, nil))
}
