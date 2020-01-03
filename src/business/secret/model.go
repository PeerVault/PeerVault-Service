// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Secret package will manage secrets
package secret

import (
	"encoding/json"
	"github.com/Power-LAB/PeerVault/database"
	"github.com/op/go-logging"
	"go.etcd.io/bbolt"
	"regexp"
)

const (
	SecretTypePassword = iota // Secret ASCII Password
	SecretTypeRsa = iota // Secret RSA key
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
	db, err := database.Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()
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

// @deprecated
func (secret *Secret) FetchSecret() error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("secret"))
		buf := b.Get([]byte("buf"))
		err := json.Unmarshal(buf, secret)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (secret *Secret) CreateSecret() error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

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
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

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