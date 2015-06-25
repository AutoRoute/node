package node

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"log"
	"math/big"
)

type PrivateKey struct {
	k *ecdsa.PrivateKey
}

func (p PrivateKey) PublicKey() PublicKey {
	return PublicKey(p.k.PublicKey)
}

func (p PrivateKey) Sign(m []byte) Signature {
	r, s, err := ecdsa.Sign(rand.Reader, p.k, m)
	if err != nil {
		log.Fatal(err)
	}
	return Signature{r, s, PublicKey(p.k.PublicKey), m}
}

type Signature struct {
	R *big.Int
	S *big.Int
	K PublicKey
	M []byte
}

func (e Signature) Key() PublicKey {
	return e.K
}

func (e Signature) Verify() error {
	k := ecdsa.PublicKey(e.K)
	if !ecdsa.Verify(&k, e.M, e.R, e.S) {
		return errors.New("Invalid Signature")
	}
	return nil
}

func (e Signature) Signed() []byte {
	return e.M
}

type PublicKey ecdsa.PublicKey

func hashstring(s string) string {
	o := sha512.Sum512([]byte(s))
	return string(o[0:sha512.Size])
}

func (e PublicKey) Hash() NodeAddress {
	// Cannot error
	t1, _ := e.X.MarshalText()
	t2, _ := e.Y.MarshalText()
	return NodeAddress(hashstring("ecdsa:P521:" + string(t1) + "," + string(t2)))
}

func NewECDSAKey() (PrivateKey, error) {
	private, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return PrivateKey{}, err
	}
	return PrivateKey{private}, nil
}
