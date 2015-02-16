package node

import (
	"github.com/AutoRoute/l2"
)

// The layer two protocol takes a layer two device and returns the hash of the
// Public Key of all neighbors it can find.
type NeighborFinder interface {
	Find(l2.FrameReadWriter) <-chan string
}

// Dummy interface to represent public keys
type PublicKey interface {
	Hash() string
}

type Packet interface {
	Destination() string
}

// A map of a fixed size representing an interfaces potential
type ReachabilityMap interface {
	IsReachable(p PublicKey)
}

// A receipt listing packets which have been succesfully delivered
type PacketReceipt interface {
	ListPackets() []string
	Verify(p PublicKey) error
}

// Layer three interfaces for network control traffic
type MapConnection interface {
	SendMap(ReachabilityMap) error
	ReadabilityMaps() <-chan ReachabilityMap
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
