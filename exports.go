package node

import (
	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

// Represents a money type which is fake and purely usable for testing
func FakeMoney() types.Money {
	return internal.FakeMoney{}
}

// Represents a money connection to an active bitcoin server
func NewRPCMoney(host, user, pass string) (types.Money, error) {
	return internal.NewRPCMoney(host, user, pass)
}

// Represents an object capable of sending and receiving packets.
type DataConnection interface {
	SendPacket(types.Packet) error
	Packets() <-chan types.Packet
}
