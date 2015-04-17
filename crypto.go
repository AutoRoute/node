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

type PrivateKey interface {
	PublicKey() PublicKey
	Sign([]byte) Signature
}

type PublicKey interface {
	Hash() NodeAddress
}

type Signature interface {
	Key() PublicKey
	Verify() error
	Signed() []byte
}

type privateECDSAEncoding struct {
	k *ecdsa.PrivateKey
}

func (p privateECDSAEncoding) PublicKey() PublicKey {
	return ecdsaEncoding(p.k.PublicKey)
}

func (p privateECDSAEncoding) Sign(m []byte) Signature {
	r, s, err := ecdsa.Sign(rand.Reader, p.k, m)
	if err != nil {
		log.Fatal(err)
	}

	return ecdsaSignature{r, s, ecdsaEncoding(p.k.PublicKey), m}
}

type ecdsaSignature struct {
	r *big.Int
	s *big.Int
	k ecdsaEncoding
	m []byte
}

func (e ecdsaSignature) Key() PublicKey {
	return e.k
}

func (e ecdsaSignature) Verify() error {
	k := ecdsa.PublicKey(e.k)
	if !ecdsa.Verify(&k, e.m, e.r, e.s) {
		return errors.New("Invalid Signature")
	}
	return nil
}

func (e ecdsaSignature) Signed() []byte {
	return e.m
}

type ecdsaEncoding ecdsa.PublicKey

func hashstring(s string) string {
	o := sha512.Sum512([]byte(s))
	return string(o[0:sha512.Size])
}

func (e ecdsaEncoding) Hash() NodeAddress {
	// Cannot error
	t1, _ := e.X.MarshalText()
	t2, _ := e.Y.MarshalText()
	return NodeAddress(hashstring("ecdsa:P521:" + string(t1) + "," + string(t2)))
}

func NewECDSAKey() (PrivateKey, error) {
	private, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return privateECDSAEncoding{private}, nil
}
