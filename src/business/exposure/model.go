// Copyright (c) 2020, Pierre Tomasina
// Use of this source code is governed by a GNU AGPLv3
// license that can be found in the LICENSE file.
//
// Exposure package will manage the secret exposure to the client
package exposure

import (
	"encoding/json"
	"github.com/Power-LAB/PeerVault/database"
	"go.etcd.io/bbolt"
	"time"
)

type Share struct {
	Uuid string
	Sender string
	Receiver string
	Expiration string
	KeyPath string
}

type ShareRequest struct {
	Receiver string
	KeyPath string
	ExpirationDelay time.Duration // Hours during the sharing request will be valid
}

type ShareResponse struct {
	Uuid string
	Sender string
	Approved bool
}

func (s *Share) Save() error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	// Share serialized to json.
	buf, err := json.Marshal(&s)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		b = tx.Bucket([]byte("share"))
		if b == nil {
			log.Debug("bucket share is nil")
			b2, err := tx.CreateBucket([]byte("share"))
			if err != nil {
				log.Debug("bucket share create error nil")
				return err
			}
			b = b2
		}
		return b.Put([]byte(s.Uuid), buf)
	})
}

func (s *Share) Delete() error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		b = tx.Bucket([]byte("share"))
		if b == nil {
			return nil
		}
		log.Debugf("Delete the share uuid: %s", s.Uuid)
		if err := b.Delete([]byte(s.Uuid)); err != nil {
			return err
		}
		return nil
	})
}

func FetchShares() ([]Share, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}
	var shares []Share

	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("share"))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var share Share
			err := json.Unmarshal(v, &share)
			if err != nil {
				return err
			}
			shares = append(shares, share)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return shares, nil
}