package node

import (
	"sync"
	"time"
)

// A Node is the highest level abstraction over the network. You receive packets
// from it and send packets to it, and it takes care of everything else.
type Node interface {
	DataConnection
}

type node struct {
	l              *sync.Mutex
	r              Router
	id             PrivateKey
	outgoing       chan Packet
	receipt_buffer []PacketHash
	receipt_ticker <-chan time.Time
	payment_ticker <-chan time.Time
}

func (n *node) receivePackets() {
	for p := range n.r.Packets() {
		n.l.Lock()
		n.receipt_buffer = append(n.receipt_buffer, p.Hash())
		n.l.Unlock()
		n.outgoing <- p
	}
}

func (n *node) sendReceipts() {
	for range n.receipt_ticker {
		n.l.Lock()
		r := CreateMerkleReceipt(n.id, n.receipt_buffer)
		n.receipt_buffer = nil
		n.l.Unlock()
		n.r.SendReceipt(r)
	}
}

func (n *node) SendPacket(p Packet) error {
	return n.r.SendPacket(p)
}

func (n *node) Packets() <-chan Packet {
	return n.outgoing
}
