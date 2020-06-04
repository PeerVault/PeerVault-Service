// Copyright (c) 2020, Pierre Tomasina
// Use of this source code is governed by a GNU AGPLv3
// license that can be found in the LICENSE file.
//
// Event package will manage all the event dispatching to
// the local client using websocket
package event

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("peerVaultLogger")
	upgrader = websocket.Upgrader{} // use default options
	connections = make([]websocket.Conn,0)
)

type Message struct {
	Type string `json:"type"`
	Data map[string]string `json:"data"`
}

// Write response to all websocket client connected
func Write(message Message) error {
	msgJson, err := json.MarshalIndent(message, "", " ")
	if err != nil {
		return err
	}
	for _, conn := range connections {
		if err := conn.WriteJSON(string(msgJson)); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

// Process a new websocket communication
func process(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	connections = append(connections, *conn)

	result := Message {
		Type: "process-ok",
		Data: nil,
	}
	resultJson, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	if err := conn.WriteJSON(string(resultJson)); err != nil {
		log.Error(err)
		return
	}

}

func Listen(address* string) {
	log.Info("listen from event")
	http.HandleFunc("/events", process)
	_ = http.ListenAndServe(*address, nil)
	select {}
}
