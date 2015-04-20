package node

import (
	"sync"
	"time"
)

// A Node is the highest level abstraction over the network. You receive packets
// from it and send packets to it, and it takes care of everything else.
type Node interface {
	AddConnection(Connection)
	DataConnection
}

type node struct {
	Router
	l              *sync.Mutex
	id             PrivateKey
	outgoing       chan Packet
	receipt_buffer []PacketHash
	receipt_ticker <-chan time.Time
	payment_ticker <-chan time.Time
	m              Money
}

func NewNode(pk PrivateKey) Node {
	n := &node{
		newRouter(pk.PublicKey()),
		&sync.Mutex{},
		pk,
		make(chan Packet),
		nil,
		time.Tick(time.Second),
		time.Tick(time.Second),
		fakeMoney{pk.PublicKey().Hash()},
	}
	go n.receivePackets()
	go n.sendReceipts()
	go n.sendPayments()
	return n
}

func (n *node) receivePackets() {
	for p := range n.Router.Packets() {
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
		n.Router.SendReceipt(r)
	}
}

func (n *node) sendPayments() {
	for range n.payment_ticker {
		//TODO
	}
}

func (n *node) SendPacket(p Packet) error {
	return n.Router.SendPacket(p)
}

func (n *node) Packets() <-chan Packet {
	return n.outgoing
}
