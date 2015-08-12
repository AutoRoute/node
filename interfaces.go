package node

import (
	"encoding/hex"
	"io"
)

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

// Layer three interfaces for network control traffic
type MapConnection interface {
	SendMap(BloomReachabilityMap) error
	ReachabilityMaps() <-chan BloomReachabilityMap
}
type ReceiptConnection interface {
	SendReceipt(PacketReceipt) error
	PacketReceipts() <-chan PacketReceipt
}
type PaymentConnection interface {
	SendPayment(PaymentHash) error
	Payments() <-chan PaymentHash
}

// While the two connections use different messages, a working ControlConnection has both interfaces
type ControlConnection interface {
	MapConnection
	ReceiptConnection
	PaymentConnection
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
