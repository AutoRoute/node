package node

type pktest string

// Dummy interface to represent public keys
type PublicKey interface {
	Hash() NodeAddress
}

func (p pktest) Hash() NodeAddress {
	return NodeAddress(p)
}
