// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Manage general database operation
package database

import (
	"encoding/binary"
	"fmt"
	"go.etcd.io/bbolt"
	"time"
)

const (
	dbfile = "peervault.db"
)

// Open PeerVault database
func Open() (*bbolt.DB, error) {
	// TODO manage path of the database
	db, err := bbolt.Open(dbfile, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	fmt.Println("After init db")
	if err != nil {
		fmt.Println("DB Not accessible, must retry later")
		return nil, err
	}
	return db, nil
}

// Convert integer into byte
func Itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}