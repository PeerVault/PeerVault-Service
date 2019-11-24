// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Secret package will manage secrets
package secret

import (
	"encoding/json"
	"fmt"

	//"github.com/Power-LAB/PeerVault/identity"
	"github.com/Power-LAB/PeerVault/database"
	"go.etcd.io/bbolt"
)

const (
	SecretTypePassword = iota // Secret ASCII Password
	SecretTypeRsa = iota // Secret RSA key
)

type Secret struct {
	// Public
	Namespace string
	Type int // SecretTypePassword | SecretTypeRsa
	Key string
	Value string
	Description string
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
			fmt.Println(secret.Key)
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
			fmt.Println("bucket secret is nil")
			b2, err := tx.CreateBucket([]byte("secret"))
			if err != nil {
				fmt.Println("bucket secret create error nil")
				return err
			}
			b = b2
		}
		return b.Put([]byte(secret.Namespace + "." + secret.Key), buf)
	})
}

func DeleteSecret(key string) error {
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
		_ = b.Delete([]byte(key))
		return nil
	})
}