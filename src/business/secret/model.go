// Package secret will manage secrets
//
// Copyright (c) 2020, Pierre Tomasina
// Use of this source code is governed by a GNU AGPLv3
// license that can be found in the LICENSE file.
package secret

import (
	"encoding/json"
	"fmt"
	"github.com/PeerVault/PeerVault-Service/database"
	"github.com/op/go-logging"
	"go.etcd.io/bbolt"
	"regexp"
)

type Error int

func (k Error) Error() (msg string) {
	switch k {
	case ErrorSecretNotFound:
		msg = "The Secret key path, namespace and key name was not found"
	}
	return fmt.Sprintf("%s (%d)", msg, k)
}

const (
	SecretTypePassword = iota // Secret ASCII Password
	SecretTypeRsa = iota // Secret RSA key

	ErrorSecretNotFound = Error(1)
)

var (
	log = logging.MustGetLogger("peerVaultLogger")
)


type Secret struct {
	// Public
	Namespace string
	Type int // SecretTypePassword | SecretTypeRsa
	Key string
	Value string
	Description string
}

func (secret *Secret) assertSecretStruct() bool {
	reNs := regexp.MustCompile("^[0-9A-Za-z_.-]+$")
	reKey := regexp.MustCompile("^[0-9A-Za-z_-]+$")

	return reNs.MatchString(secret.Namespace) && reKey.MatchString(secret.Key) && secret.Type <= SecretTypeRsa
}

func FetchSecrets() ([]Secret, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	var secrets []Secret

	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("secret"))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var secret Secret
			err := json.Unmarshal(v, &secret)
			if err != nil {
				return err
			}
			log.Debug(secret.Key)
			log.Debug(secret.Value)
			secret.Value = ""
			secrets = append(secrets, secret)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

func FetchSecret(keyPath []byte) (Secret, error) {
	secret := &Secret{}
	db, err := database.GetConnection()
	if err != nil {
		return *secret, err
	}

	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("secret"))
		if b == nil {
			return ErrorSecretNotFound
		}
		buf := b.Get(keyPath)
		if buf == nil {
			return ErrorSecretNotFound
		}
		return json.Unmarshal(buf, secret)
	})
	if err != nil {
		return *secret, err
	}

	return *secret, nil
}

func (secret *Secret) CreateSecret() error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	// Secret serialized to json.
	buf, err := json.Marshal(&secret)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		b = tx.Bucket([]byte("secret"))
		if b == nil {
			log.Debug("bucket secret is nil")
			b2, err := tx.CreateBucket([]byte("secret"))
			if err != nil {
				log.Debug("bucket secret create error nil")
				return err
			}
			b = b2
		}
		return b.Put([]byte(secret.Namespace + "." + secret.Key), buf)
	})
}

// keyPath are the fullpath of the key, namespace concat with key, spaced by dot
func DeleteSecret(keyPath string) error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		b = tx.Bucket([]byte("secret"))
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