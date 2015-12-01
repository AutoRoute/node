package node

import (
	"bytes"
	"math/big"
	"testing"
)

func TestSigning(t *testing.T) {
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

func TestBadSignature(t *testing.T) {
	var sig Signature
	if sig.Verify() == nil {
		t.Fatal("Empty signature should fail to verify")
	}

	k, err := NewECDSAKey()
	if err != nil {
		t.Fatal(err)
	}
	m := []byte("hello")
	sig = k.Sign(m)

	sig.M = nil
	if sig.Verify() == nil {
		t.Fatal("Empty M should fail to verify")
	}

	sig = k.Sign(m)
	sig.M = big.NewInt(1).Bytes()
	if sig.Verify() == nil {
		t.Fatal("Empty M should fail to verify")
	}
}

func TestBadCryptoMarshaling(t *testing.T) {
	k := &PrivateKey{}
	if k.UnmarshalJSON([]byte("BAD JSON")) == nil {
		t.Fatal("Expected failure to unmarshal private key")
	}

	pk := &PublicKey{}
	if pk.UnmarshalJSON([]byte("BAD JSON")) == nil {
		t.Fatal("Expected failure to unmarshal public key")
	}
}
