package types

// This package contains types which are publicly exported, but are used by internal libraries.

import (
	"encoding/hex"
)

// This represents a raw node address (not hex encoded for human consumption).
type NodeAddress string

func (n NodeAddress) MarshalText() ([]byte, error) {
	s := hex.EncodeToString([]byte(n))
	return []byte(s), nil
}
func (n *NodeAddress) UnmarshalText(b []byte) error {
	s, err := hex.DecodeString(string(b))
	*n = NodeAddress(string(s))
	return err
}

// This represents a canonical packet hash.
type PacketHash string

func (n PacketHash) MarshalText() ([]byte, error) {
	s := hex.EncodeToString([]byte(n))
	return []byte(s), nil
}
func (n *PacketHash) UnmarshalText(b []byte) error {
	s, err := hex.DecodeString(string(b))
	*n = PacketHash(string(s))
	return err
}

// This represents a payment engine, which produces signed payments on demand
type Money interface {
	MakePayment(amount int64, destination string) (chan bool, error)
	GetNewAddress() (string, chan uint64, error)
}
