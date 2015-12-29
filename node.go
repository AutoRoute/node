package node

import (
	"log"
	"sync"
	"time"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

// A Node is the highest level abstraction over the network. You receive packets
// from it and send packets to it, and it takes care of everything else.
type Node struct {
	private *fullNode
}

func (n Node) IsReachable(addr types.NodeAddress) bool {
	return n.private.IsReachable(addr)
}

func (n Node) SendPacket(p types.Packet) error {
	return n.private.SendPacket(p)
}

func (n Node) Packets() <-chan types.Packet {
	return n.private.Packets()
}

// A fullNode includes various functions which are called by Server but shouldn't be publicly exposed.
type fullNode struct {
	router         *internal.Router
	l              *sync.Mutex
	id             internal.PrivateKey
	outgoing       chan types.Packet
	receipt_buffer []types.PacketHash
	receipt_ticker <-chan time.Time
	payment_ticker <-chan time.Time
	m              types.Money
	quit           chan bool
}

func newFullNode(pk internal.PrivateKey, m types.Money, receipt_ticker <-chan time.Time, payment_ticker <-chan time.Time) *fullNode {
	n := &fullNode{
		internal.NewRouter(pk.PublicKey()),
		&sync.Mutex{},
		pk,
		make(chan types.Packet),
		nil,
		receipt_ticker,
		payment_ticker,
		m,
		make(chan bool),
	}
	go n.receivePackets()
	go n.sendReceipts()
	go n.sendPayments()
	return n
}

func (n *fullNode) receivePackets() {
	for {
		select {
		case p, ok := <-n.router.Packets():
			if !ok {
				return
			}
			n.l.Lock()
			n.receipt_buffer = append(n.receipt_buffer, p.Hash())
			n.l.Unlock()
			n.outgoing <- p
		case <-n.quit:
			return
		}
	}
}

func (n *fullNode) sendReceipts() {
	for {
		select {
		case <-n.receipt_ticker:
			n.l.Lock()
			if len(n.receipt_buffer) > 0 {
				r := internal.CreateMerkleReceipt(n.id, n.receipt_buffer)
				n.receipt_buffer = nil
				n.router.SendReceipt(r)
			}
			n.l.Unlock()
		case <-n.quit:
			return
		}
	}
}

func (n *fullNode) sendPayments() {
	for {
		select {
		case <-n.payment_ticker:
			n.l.Lock()
			for _, c := range n.router.Connections() {
				owed := n.router.OutgoingDebt(c.Key().Hash())
				log.Printf("Owe %x %d", c.Key().Hash(), owed)
				if owed > 0 {
					log.Printf("Sending payment to %s", c.OtherMetaData().Payment_Address)
					p, err := n.m.MakePayment(owed, c.OtherMetaData().Payment_Address)
					if err != nil {
						log.Printf("Failed to make a payment to %x (%x) : %v",
							c.Key().Hash(), c.OtherMetaData().Payment_Address, err)
						break
					}
					go n.router.RecordPayment(c.Key().Hash(), owed, p)
				}
			}
			n.l.Unlock()
		case <-n.quit:
			return
		}
	}
}

func (n *fullNode) SendPacket(p types.Packet) error {
	return n.router.SendPacket(p)
}

func (n *fullNode) Packets() <-chan types.Packet {
	return n.outgoing
}

func (n *fullNode) Close() error {
	close(n.quit)
	return nil
}

func (n *fullNode) GetNewAddress() string {
	address, c, err := n.m.GetNewAddress()
	if err != nil {
		log.Fatal("Failed to get payment address: ", err)
	}
	n.router.Ledger.AddAddress(address, c)
	return address
}

func (n *fullNode) GetAddress() internal.PublicKey {
	return n.id.PublicKey()
}

func (n *fullNode) IsReachable(addr types.NodeAddress) bool {
	_, err := n.router.FindNextHop(addr)
	return err == nil
}

func (n *fullNode) AddConnection(c internal.Connection) {
	n.router.AddConnection(c)
}
