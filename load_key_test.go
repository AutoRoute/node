package node

import (
	"os"
	"testing"
)

func TestCreateKey(t *testing.T) {
	keyfile := "test_keyfile"
	defer os.Remove(keyfile)
	f, k1, err := CreateKey(keyfile)
	if err != nil {
		t.Fatal(err)
	}
	k2, err := LoadKey(keyfile)
	if err != nil {
		t.Fatal(err)
	}
	if k1.PublicKey().Hash() != k2.k.PublicKey().Hash() {
		t.Fatalf("diff public keys %v != %v", k1, k2)
	}
	if k1.K.D.Cmp(k2.k.K.D) != 0 {
		t.Fatalf("diff D %v != %v", k1, k2)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	k3, err := LoadKey(keyfile)
	if err != nil {
		t.Fatal(err)
	}
	k4, err := LoadKey(keyfile)
	if err != nil {
		t.Fatal(err)
	}
	if k3.k.PublicKey().Hash() != k4.k.PublicKey().Hash() {
		t.Fatalf("diff public keys %v != %v", k1, k2)
	}
	if k3.k.K.D.Cmp(k4.k.K.D) != 0 {
		t.Fatalf("diff D %v != %v", k1, k2)
	}
}

func TestLoadKey(t *testing.T) {
	keyfile := "test_keyfile"
	defer os.Remove(keyfile)
	k1, err := LoadKey(keyfile)
	if err != nil {
		t.Fatal(err)
	}
	k2, err := LoadKey(keyfile)
	if err != nil {
		t.Fatal(err)
	}
	if k1.k.PublicKey().Hash() != k2.k.PublicKey().Hash() {
		t.Fatalf("diff public keys %v != %v", k1, k2)
	}
	if k1.k.K.D.Cmp(k2.k.K.D) != 0 {
		t.Fatalf("diff D %v != %v", k1, k2)
	}
}
