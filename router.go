package node

import (
	"time"
)

// A router handles all routing tasks that don't involve the local machine
// including connection management, reachability handling, packet receipt
// relaying, and (outstanding) payment tracking. AKA anything which doesn't
// need the private key.
type Router interface {
	AddConnection(Connection)
	DataConnection
	GetAddress() PublicKey
	SendReceipt(PacketReceipt)
	SendPayment(NodeAddress, PaymentHash) error
	RecordPayment(Payment)
	Connections() []NodeAddress
	IncomingDebt(NodeAddress) (int64, time.Time)
	OutgoingDebt(NodeAddress) (int64, time.Time)
}

type router struct {
	pk PublicKey
	// A map of public key hashes to connections
	connections map[NodeAddress]Connection

	ReachabilityHandler
	RoutingHandler
	ReceiptHandler
	PaymentHandler
	Ledger
}

func newRouter(pk PublicKey) Router {
	reach := newReachability(pk.Hash())
	routing := newRouting(pk, reach)
	c1, c2 := SplitChannel(routing.Routes())
	receipt := newReceipt(pk.Hash(), c1)
	payment := newPayment(pk.Hash())
	ledger := newLedger(pk.Hash(), receipt.PacketHashes(), c2)
	return &router{
		pk,
		make(map[NodeAddress]Connection),
		reach,
		routing,
		receipt,
		payment,
		ledger}
}

func (r *router) GetAddress() PublicKey {
	return r.pk
}

func (r *router) Connections() []NodeAddress {
	c := make([]NodeAddress, 0, len(r.connections))
	for k, _ := range r.connections {
		c = append(c, k)
	}
	return c
}

func (r *router) AddConnection(c Connection) {
	id := c.Key().Hash()
	_, duplicate := r.connections[id]
	if duplicate {
		c.Close()
		return
	}
	r.connections[id] = c

	// Curry the id since the various sub connections don't know about it
	r.RoutingHandler.AddConnection(id, c)
	r.ReachabilityHandler.AddConnection(id, c)
	r.ReceiptHandler.AddConnection(id, c)
	r.PaymentHandler.AddConnection(id, c)
}
