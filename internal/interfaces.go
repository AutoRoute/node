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

// A simple interface for something that uses an algorithm decide where to send
// packets.
type routingAlgorithm interface {
	// Finds the next place to send a packet.
	// Args:
	//  id: The destination node.
	//  src: The source node. (So we don't send it backwards.)
	// Returns:
	//  The next node that we should send the packet to, error
	FindNextHop(id types.NodeAddress,
		src types.NodeAddress) (types.NodeAddress, error)
	// Sets the routingHandler that will be used with this routingAlgorithm. This
	// should be called by the routingHandler upon initialization. A routing
	// algorithm will generally require this method to be called before it can be
	// used.
	// Args:
	//  routing: The routingHandler to use.
	BindToRouting(routing *routingHandler)
	// Does any required cleanup.
	Cleanup()
}

// Interface for something that can log routing decisions
type Logger interface {
	LogBloomFilter(*BloomReachabilityMap) error
	LogRoutingDecision(types.NodeAddress, types.NodeAddress, int, int64, types.PacketHash) error
}
