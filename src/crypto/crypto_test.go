package crypto

import (
  "testing"
  "strings"
)

func TestCreateMnemonic(t *testing.T) {
  seed := &Seed {}
  seed.CreateMnemonic()

  if len(strings.Fields(seed.Mnemonic)) != 24 {
    t.Errorf("mnemonic seed must be 24 words, current: %s", seed.Mnemonic)
  }
}

func TestCreateMasterKey(t *testing.T) {
  t.Skip("TODO TestCreateMasterKey")
}

func TestCreateChildKey(t *testing.T) {
  t.Skip("TODO TestCreateChildKey")
}

func TestIsChildFromMaster(t *testing.T) {
  t.Skip("TODO TestIsChildFromMaster")
}

func TestBipKeyToLibp2p(t *testing.T) {
  t.Skip("TODO TestBipKeyToLibp2p")
}
