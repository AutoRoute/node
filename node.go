package node

import (
	"log"
	"sync"
	"time"
)

// A Node is the highest level abstraction over the network. You receive packets
// from it and send packets to it, and it takes care of everything else.
type Node struct {
	*router
	l              *sync.Mutex
	id             PrivateKey
	outgoing       chan Packet
	receipt_buffer []PacketHash
	receipt_ticker <-chan time.Time
	payment_ticker <-chan time.Time
	m              Money
	quit           chan bool
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
		make(chan bool),
	}
	go n.receivePackets()
	go n.sendReceipts()
	go n.sendPayments()
	return n
}

func (n *Node) receivePackets() {
	for {
		select {
		case p := <-n.router.Packets():
			n.l.Lock()
			n.receipt_buffer = append(n.receipt_buffer, p.Hash())
			n.l.Unlock()
			n.outgoing <- p
		case <-n.quit:
			return
		}
	}
}

func (n *Node) sendReceipts() {
	for {
		select {
		case <-n.receipt_ticker:
			n.l.Lock()
			if len(n.receipt_buffer) > 0 {
				r := CreateMerkleReceipt(n.id, n.receipt_buffer)
				n.receipt_buffer = nil
				n.router.SendReceipt(r)
			}
			n.l.Unlock()
		case <-n.quit:
			return
		}
	}
}

func (n *Node) sendPayments() {
	for {
		select {
		case <-n.payment_ticker:
			n.l.Lock()
			for _, c := range n.router.Connections() {
				owed, _ := n.router.OutgoingDebt(c)
				if owed > 0 {
					p, err := n.m.MakePayment(owed, c)
					if err != nil {
						log.Printf("Failed to make a payment to %s : %v", c, err)
						break
					}
					n.router.RecordPayment(Payment{n.id.PublicKey().Hash(), c, owed})
					n.router.SendPaymentHash(c, p)
				}
			}
			n.l.Unlock()
		case <-n.quit:
			return
		}
	}
}

func (n *Node) receivePayments() {
	for {
		select {
		case h := <-n.router.PaymentHashes():
			n.l.Lock()
			c := n.m.AddPaymentHash(h)
			go func() {
				select {
				case p := <-c:
					n.router.RecordPayment(p)
				case <-n.quit:
					return
				}
			}()
		case <-n.quit:
			return
		}
	}
}

func (n *Node) SendPacket(p Packet) error {
	return n.router.SendPacket(p)
}

func (n *Node) Packets() <-chan Packet {
	return n.outgoing
}

func (n *Node) Close() error {
	close(n.quit)
	return nil
}
