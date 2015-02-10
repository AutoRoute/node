package node

import (
	"github.com/AutoRoute/l2"
)

// The layer two protocol takes a layer two device and returns the hash of the
// Public Key of all neighbors it can find.
type NeighborFinder interface {
	Find(l2.FrameReadWriter) <-chan string
}

// A map of a fixed size representing an interfaces potential
type ReachabilityMap interface{}

// A receipt listing packets which have been succesfully deliver.
type PacketReceipt interface{}

// Layer three interfaces for network control traffic
type MapConnection interface {
	SendMap(ReachabilityMap)
	ReadabilityMaps() <-chan ReachabilityMap
}
type ReceiptConnection interface {
	SendReceipt(PacketReceipt)
	PacketReceipts() <-chan PacketReceipt
}

// While the two connections use different messages, a working ControlConnection has both interfaces
type ControlConnection interface {
	MapConnection
	ReceiptConnection
}

// The actual data connection. Should be done at the layer two level in order to be able to send congestion signals
type DataConnection interface {
	SendPacket([]byte)
	Packets() <-chan []byte
}
