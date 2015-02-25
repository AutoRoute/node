package node

import ()

type pktest string

// Dummy interface to represent public keys
type PublicKey interface {
	Hash(p pktest) NodeAddress
}

func Hash(p pktest) NodeAddress {
	return NodeAddress(p)
}
