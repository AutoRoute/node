package node

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

// Dummy interface to represent public keys
type PublicKey interface {
	Hash() NodeAddress
}

type ecdsaEncoding ecdsa.PublicKey

func (e ecdsaEncoding) Hash() NodeAddress {
	t1, err := e.X.MarshalText()
	if err != nil {
		panic(err)
	}
	t2, err := e.Y.MarshalText()
	if err != nil {
		panic(err)
	}
	return NodeAddress("ecdsa:P521:" + string(t1) + "," + string(t2))
}

func NewPublicKey() (PublicKey, error) {
	private, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return ecdsaEncoding(private.PublicKey), nil
}
