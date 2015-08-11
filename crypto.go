package node

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"log"
	"math/big"
)

type PrivateKey struct {
	k *ecdsa.PrivateKey
}

type encodedprivk struct {
	D  *big.Int
	PK PublicKey
}

func (e PrivateKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(encodedprivk{D: e.k.D, PK: e.PublicKey()})
}

func (e *PrivateKey) UnmarshalJSON(b []byte) error {
	var s encodedprivk
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	e.k = &ecdsa.PrivateKey{ecdsa.PublicKey(s.PK), s.D}
	return nil
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

type encodedpk struct {
	X *big.Int
	Y *big.Int
}

func (e PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(encodedpk{X: e.X, Y: e.Y})
}

func (e *PublicKey) UnmarshalJSON(b []byte) error {
	var s encodedpk
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	e.X = s.X
	e.Y = s.Y
	e.Curve = elliptic.P521()
	return nil
}

func NewECDSAKey() (PrivateKey, error) {
	private, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return PrivateKey{}, err
	}
	return PrivateKey{private}, nil
}
