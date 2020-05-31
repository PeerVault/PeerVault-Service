// +build linux

// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Crypto package will manage the cryptography of the Vault
// - Store and retrieve private key from OSX Keychain
package crypto

import (
	"fmt"
	"github.com/Power-LAB/PeerVault/database"
	"github.com/op/go-logging"
	"go.etcd.io/bbolt"
)

func EnableDevMode() {
	log.Debug("Developer mode enabled using linux secret service 'peervault-dev'")
}

type Keychain struct {
	keychain string
}

func (k *Keychain) CreateOrOpen() error {
	fmt.Printf(
		"\033[%dm%s\033[0m",
		int(logging.ColorRed),
		"!!! ATTENTION !!!\nTHE LINUX VERSION DOES NOT STORE SECURELY THE IDENTITY KEYS AS MACOS USING KEYCHAIN.\n" +
			"THE LINUX VERSION IS NOT PRODUCTION READY\n\n")
	return nil
}

func (k *Keychain) Put(key string, value []byte, label string, forceUpdate bool) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	keyPath := []byte(key + "." + label)

	return db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		b = tx.Bucket([]byte("keychain"))
		if b == nil {
			log.Debug("keychain bucket is nil")
			b2, err := tx.CreateBucket([]byte("keychain"))
			if err != nil {
				log.Debug("bucket keychain create error nil")
				return err
			}
			b = b2
		}
		return b.Put(keyPath, value)
	})
}

func (k *Keychain) Get(key string, label string) ([]byte, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	keyPath := []byte(key + "." + label)
	var value []byte
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("keychain"))
		if b == nil {
			return ErrorKeychainKeyNotFound
		}
		// Strange BUG when using direct byte result signal SIGSEGV: segmentation violation
		value = []byte(string(b.Get(keyPath)))
		if value == nil {
			return ErrorKeychainKeyNotFound
		}
		return nil
	})
	return value, err
}

func (k *Keychain) Delete(keyPath string) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		b = tx.Bucket([]byte("keychain"))
		if b == nil {
			return nil
		}
		log.Debugf("Delete the key path: %s", keyPath)
		if err := b.Delete([]byte(keyPath)); err != nil {
			return err
		}
		return nil
	})
}