// Copyright (c) 2019, Pierre Tomasina
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Crypto package will manage the cryptography of the Vault
//  - Seed based on BIP32
//  - Master key derive from seed
//  - Child key used for libp2p
//  - Local file encrypted used with master key
package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"io"

	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/op/go-logging"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
)

var (
	log = logging.MustGetLogger("peerVaultLogger")
)

//
// Seed Integration
//
type Seed struct {
	Mnemonic string
	seed     []byte
}

// Create a phrase consisting of the mnemonic words
func (s *Seed) CreateMnemonic() {
	// Generate a mnemonic for memorization or user-friendly seeds
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	s.Mnemonic = mnemonic
}

// Create a bip39 seed
// Create s.mnemonic if is not specified
func (s *Seed) CreateSeed() {
	if s.Mnemonic == "" {
		s.CreateMnemonic()
	}
	s.seed = bip39.NewSeed(s.Mnemonic, "")
}

//
// Master / Child key integration
//

// Create master private key from seed
func (s *Seed) CreateMasterKey() (*bip32.Key, error) {
	if len(s.seed) == 0 {
		return nil, errors.New("You must create seed before master key")
	}
	return bip32.NewMasterKey(s.seed)
}

// Create child key
// Generate a random index between 0 and bip32.FirstHardenedChild
// Child key must generate non-hardened key in order to prove relation with parent key
func CreateChildKey(master *bip32.Key) (*bip32.Key, error) {
	var src cryptoSource
	rnd := rand.New(src)
	childPosition := uint32(rnd.Intn(int(bip32.FirstHardenedChild)))
	// Create peer key
	return master.NewChildKey(childPosition)
}

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}
func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}
func (s cryptoSource) Uint64() (v uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func IsChildFromMaster(child *bip32.Key, master *bip32.Key) bool {
	/*
	   HMAC-SHA512(Key = cpar, Data = serP(Kpar) || ser32(i))
	   Split I into two 32-byte sequences, IL and IR.
	   The returned child key Ki is point(parse256(IL)) + Kpar.
	   The returned chain code ci is IR.
	   In case parse256(IL) ≥ n or Ki is the point at infinity, the resulting key is invalid, and one should proceed with the next value for i.
	*/
	parentFP, _ := hash160(master.PublicKey().Key)
	return bytes.Equal(parentFP[:4], child.PublicKey().FingerPrint)
}

// Convert a Secp256 BIP32 Child key intp Crypto Libp2p PrivKey struct
// Node key also known as Device Key will be used to identify a specific peer
func BipKeyToLibp2p(child *bip32.Key) (crypto.PrivKey, error) {
	return crypto.UnmarshalSecp256k1PrivateKey(child.Key)
}

// Symmetric encryption using Child key from bip32.Key
// The output are base64 encoded
func EncryptAes(key []byte, in []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Debugf("NewCipher(%d bytes)", len(key))
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Debug("GCM Galois/Counter mode fail", err)
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(crand.Reader, nonce); err != nil {
		log.Debug("Nonce random fail", err)
		return nil, err
	}
	cipherText := gcm.Seal(nonce, nonce, in, nil)
	return []byte(base64.StdEncoding.EncodeToString(cipherText)), nil
}

// Symmetric decryption using Child key from bip32.Key
// The input must be base64 encoded, returned by function EncryptAes
func DecryptAes(key []byte, in []byte) ([]byte, error) {
	cipherText, err := base64.StdEncoding.DecodeString(string(in))
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Error("NewCipher(%d bytes) = %s", len(key), err)
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Debug("GCM Galois/Counter mode fail", err)
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, errors.New("Ciphertext block size is too short!")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		log.Debug("Fail to decrypt", err)
		return nil, err
	}
	return plainText, nil
}

//
// Hashes
//
func hashSha256(data []byte) ([]byte, error) {
	hasher := sha256.New()
	_, err := hasher.Write(data)
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func hashDoubleSha256(data []byte) ([]byte, error) {
	hash1, err := hashSha256(data)
	if err != nil {
		return nil, err
	}

	hash2, err := hashSha256(hash1)
	if err != nil {
		return nil, err
	}
	return hash2, nil
}

func hashRipeMD160(data []byte) ([]byte, error) {
	hasher := ripemd160.New()
	_, err := io.WriteString(hasher, string(data))
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func hash160(data []byte) ([]byte, error) {
	hash1, err := hashSha256(data)
	if err != nil {
		return nil, err
	}

	hash2, err := hashRipeMD160(hash1)
	if err != nil {
		return nil, err
	}

	return hash2, nil
}
