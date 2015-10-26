package node

import (
	"bytes"
	"testing"
)

func TestCrypto(t *testing.T) {
	k, err := NewECDSAKey()
	if err != nil {
		t.Fatal(err)
	}
	m := []byte("hello")
	s := k.Sign(m)
	if s.Key().Hash() != k.PublicKey().Hash() {
		t.Fatalf("Expected %v != %v", s.Key().Hash(), k.PublicKey().Hash())
	}
	if s.Verify() != nil {
		t.Fatal(s.Verify())
	}
	if !bytes.Equal(s.Signed(), m) {
		t.Fatalf("Expected %v != %v", s.Signed(), m)
	}
}

func TestEmptyVerify(t *testing.T) {
	var sig Signature
	if sig.Verify() == nil {
		t.Fatal("Empty signature should fail to verify")
	}
}
