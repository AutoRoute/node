package node

import (
	"io"
)

type NodeAddress string
type PacketHash string

type Packet interface {
	Destination() NodeAddress
	Hash() PacketHash
}

// A map of a fixed size representing an interfaces potential
type ReachabilityMap interface {
	IsReachable(s NodeAddress) bool
	AddEntry(n NodeAddress)
	Increment()
	Merge(n ReachabilityMap)
}

// A receipt listing packets which have been succesfully delivered
type PacketReceipt interface {
	ListPackets() []PacketHash
	Source() NodeAddress
	Verify() error
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
	io.Closer
}
