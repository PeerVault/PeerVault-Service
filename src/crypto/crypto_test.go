package crypto

import (
	"strings"
	"testing"
)

func TestCreateMnemonic(t *testing.T) {
	seed := &Seed{}
	seed.CreateMnemonic()

	if len(strings.Fields(seed.Mnemonic)) != 24 {
		t.Errorf("mnemonic seed must be 24 words, current: %s", seed.Mnemonic)
	}
}

func TestCreateMasterKey(t *testing.T) {
	seed := &Seed{}
	seed.CreateSeed()
	master, err := seed.CreateMasterKey()

	if master == nil {
		t.Errorf("Master key fail to create, %s", err.Error())
	}
}

func TestCreateChildKey(t *testing.T) {
	seed := &Seed{}
	seed.CreateSeed()
	master, _ := seed.CreateMasterKey()
	child, err := CreateChildKey(master)

	if child == nil {
		t.Errorf("Child key fail to create, %s", err.Error())
	}
}

func TestIsChildFromMaster(t *testing.T) {
	seed := &Seed{}
	seed.CreateSeed()
	master, _ := seed.CreateMasterKey()
	child, err := CreateChildKey(master)

	if child == nil {
		t.Errorf("Child key fail to create, %s", err.Error())
	}

	if !IsChildFromMaster(child, master) {
		t.Errorf("Error, child should be from master")
	}

	seed2 := &Seed{}
	seed2.CreateSeed()
	master2, _ := seed2.CreateMasterKey()

	if IsChildFromMaster(child, master2) {
		t.Errorf("Error, child should NOT be from master2")
	}
}

func TestBipKeyToLibp2p(t *testing.T) {
	seed := &Seed{}
	seed.CreateSeed()
	master, _ := seed.CreateMasterKey()
	child, err := CreateChildKey(master)

	if child == nil {
		t.Errorf("Child key fail to create, %s", err.Error())
	}

	_, err = BipKeyToLibp2p(child)

	if err != nil {
		t.Errorf("Node key convertion fail, %s", err.Error())
	}
}
