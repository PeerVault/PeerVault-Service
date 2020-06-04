// +build darwin

// Copyright (c) 2020, Pierre Tomasina
// Use of this source code is governed by a GNU AGPLv3
// license that can be found in the LICENSE file.
//
// Crypto package will manage the cryptography of the Vault
// - Store and retrieve private key from OSX Keychain
package crypto

import (
	"errors"
	"github.com/keybase/go-keychain"
)

const (
	service = "PeerVault"
)

var (
	path = "peervault.keychain"
)

func EnableDevMode() {
	log.Debug("Developer mode enabled using keychain 'peervault-dev'")
	path = "peervault-dev.keychain"
}

type Keychain struct {
	keychain keychain.Keychain
}

func (k *Keychain) CreateOrOpen() error {
	kc := keychain.NewWithPath(path)

	log.Debug("Checking keychain status")
	err := kc.Status()
	if err == nil {
		log.Debug("Keychain status returned nil, keychain exists")
		k.keychain = kc
		return nil
	}

	log.Debug("Keychain status returned error", err)

	if err != keychain.ErrorNoSuchKeychain {
		return err
	}

	k.keychain, err = keychain.NewKeychainWithPrompt(path)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keychain) Put(key string, value []byte, label string, forceUpdate bool) error {
	if key == label {
		return errors.New("keychain key and label must be different")
	}
	var err error
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetAccount(key)
	item.SetService(service)
	item.SetLabel(label)
	item.SetData(value)
	item.UseKeychain(k.keychain)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	err = keychain.AddItem(item)

	if err == keychain.ErrorDuplicateItem {
		if false == forceUpdate {
			return ErrorKeychainValueAlreadyExists
		}
		query := keychain.NewItem()
		query.SetSecClass(keychain.SecClassGenericPassword)
		query.SetAccount(key)
		query.SetService(service)
		query.SetLabel(label)
		query.SetMatchLimit(keychain.MatchLimitOne)
		query.SetReturnData(true)
		query.SetMatchSearchList(k.keychain)
		err = keychain.UpdateItem(query, item)
	}
	if err != nil {
		log.Error("keychain put error", err)
	}
	return nil
}

func (k *Keychain) Get(key string, label string) ([]byte, error) {
	log.Debug(key, label)
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetAccount(key)
	query.SetService(service)
	query.SetLabel(label)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)
	query.SetMatchSearchList(k.keychain)

	results, err := keychain.QueryItem(query)

	if err == keychain.ErrorItemNotFound || len(results) == 0 {
		log.Error("keychain get no results found error", err)
		return nil, ErrorKeychainKeyNotFound
	}

	return results[0].Data, nil
}

func (k *Keychain) Delete(key string) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetAccount(key)
	item.SetService(service)
	item.SetMatchLimit(keychain.MatchLimitOne)
	item.SetReturnAttributes(true)
	item.SetReturnData(true)
	item.SetMatchSearchList(k.keychain)

	return keychain.DeleteItem(item)
}