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

// A Node is the highest level abstraction over the network. You receive packets
// from it and send packets to it, and it takes care of everything else.
type Node struct {
	private *internal.Node
}

func (n Node) IsReachable(addr types.NodeAddress) bool {
	return n.private.IsReachable(addr)
}

func (n Node) SendPacket(p types.Packet) error {
	return n.private.SendPacket(p)
}

func (n Node) Packets() <-chan types.Packet {
	return n.private.Packets()
}

func (n Node) GetNodeAddress() types.NodeAddress {
	return n.GetNodeAddress()
}
