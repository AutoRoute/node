package node

import (
	"github.com/AutoRoute/l2"
)

type NodeAddress string

// Dummy interface to represent public keys
type PublicKey interface {
	Hash() NodeAddress
}

type Packet interface {
	Destination() NodeAddress
}

// A map of a fixed size representing an interfaces potential
type ReachabilityMap interface {
	IsReachable(s NodeAddress) bool
}

// A receipt listing packets which have been succesfully delivered
type PacketReceipt interface {
	ListPackets() []string
	Verify(p NodeAddress) error
}

// Layer three interfaces for network control traffic
type MapConnection interface {
	SendMap(ReachabilityMap) error
	ReachabilityMaps() <-chan ReachabilityMap
}
type ReceiptConnection interface {
	SendReceipt(PacketReceipt) error
	PacketReceipts() <-chan PacketReceipt
}

// While the two connections use different messages, a working ControlConnection has both interfaces
type ControlConnection interface {
	MapConnection
	ReceiptConnection
}

// The actual data connection. Should be done at the layer two level in order to be able to send congestion signals
type DataConnection interface {
	SendPacket(Packet) error
	Packets() <-chan Packet
}

type Connection interface {
	ControlConnection
	DataConnection
	Key() PublicKey
	Close()
}
