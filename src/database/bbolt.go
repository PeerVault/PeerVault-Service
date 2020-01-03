// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Manage general database operation
package database

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"time"
)

var (
	dbFilePath string
)

func SetDbPath(path string) {
	dbFilePath = path
}

func GetDbPath() string {
	return dbFilePath
}

// Open PeerVault database
func Open() (*bbolt.DB, error) {
	_ = os.MkdirAll(filepath.Dir(dbFilePath), 0700)
	db, err := bbolt.Open(dbFilePath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
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

func StructToBytes(v interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func BytesToStruct(data []byte, v interface{}) error {
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, &v)
	if err != nil {
		panic(err)
	}
	return nil
}