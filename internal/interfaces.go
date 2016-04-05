package internal

import (
	"io"

	"github.com/AutoRoute/node/types"
)

// Layer three interfaces for network control traffic
type MapConnection interface {
	SendMap(*BloomReachabilityMap) error
	ReachabilityMaps() <-chan *BloomReachabilityMap
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
	SendPacket(types.Packet) error
	Packets() <-chan types.Packet
}

type Connection interface {
	ControlConnection
	DataConnection
	Key() PublicKey
	// My sides metadata
	MetaData() SSHMetaData
	OtherMetaData() SSHMetaData
	io.Closer
}

// A simple interface for something that uses an algorithm to route packets.
type routingHandler interface {
  // Register a connection to another node.
  AddConnection(id types.NodeAddress, c DataConnection)
  // Route a new packet.
  SendPacket(p types.Packet) error

  // Get all the routes we know of.
  Routes() <-chan routingDecision
  Packets() <-chan types.Packet
  Close() error
}
