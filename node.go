package node

import (
	"log"
	"sync"
	"time"
)

// A Node is the highest level abstraction over the network. You receive packets
// from it and send packets to it, and it takes care of everything else.
type Node struct {
	Router
	l              *sync.Mutex
	id             PrivateKey
	outgoing       chan Packet
	receipt_buffer []PacketHash
	receipt_ticker <-chan time.Time
	payment_ticker <-chan time.Time
	m              Money
}

func NewNode(pk PrivateKey, receipt_ticker <-chan time.Time, payment_ticker <-chan time.Time) *Node {
	n := &Node{
		newRouter(pk.PublicKey()),
		&sync.Mutex{},
		pk,
		make(chan Packet),
		nil,
		receipt_ticker,
		payment_ticker,
		fakeMoney{pk.PublicKey().Hash()},
	}
	go n.receivePackets()
	go n.sendReceipts()
	go n.sendPayments()
	return n
}

func (n *Node) receivePackets() {
	for p := range n.Router.Packets() {
		n.l.Lock()
		n.receipt_buffer = append(n.receipt_buffer, p.Hash())
		n.l.Unlock()
		n.outgoing <- p
	}
}

func (n *Node) sendReceipts() {
	for range n.receipt_ticker {
		n.l.Lock()
		if len(n.receipt_buffer) > 0 {
			r := CreateMerkleReceipt(n.id, n.receipt_buffer)
			n.receipt_buffer = nil
			n.Router.SendReceipt(r)
		}
		n.l.Unlock()
	}
}

func (n *Node) sendPayments() {
	for range n.payment_ticker {
		n.l.Lock()
		for _, c := range n.Router.Connections() {
			owed, _ := n.Router.OutgoingDebt(c)
			if owed > 0 {
				p, err := n.m.MakePayment(owed, c)
				if err != nil {
					log.Printf("Failed to make a payment to %s : %v", c, err)
					break
				}
				n.Router.RecordPayment(Payment{n.id.PublicKey().Hash(), c, owed})
				n.Router.SendPaymentHash(c, p)
			}
		}
		n.l.Unlock()
	}
}

func (n *Node) receivePayments() {
	for h := range n.Router.PaymentHashes() {
		n.l.Lock()
		c := n.m.AddPaymentHash(h)
		go func() {
			p := <-c
			n.Router.RecordPayment(p)
		}()
	}
}

func (n *Node) SendPacket(p Packet) error {
	return n.Router.SendPacket(p)
}

func (n *Node) Packets() <-chan Packet {
	return n.outgoing
}
