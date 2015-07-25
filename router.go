package node

import (
	"sync"
)

// A router handles all routing tasks that don't involve the local machine
// including connection management, reachability handling, packet receipt
// relaying, and (outstanding) payment tracking. AKA anything which doesn't
// need the private key.
type router struct {
	pk PublicKey
	// A map of public key hashes to connections
	connections map[NodeAddress]Connection

	*reachabilityHandler
	*routingHandler
	*receiptHandler
	*paymentHandler
	*ledger

	lock *sync.Mutex
}

func newRouter(pk PublicKey) *router {
	reach := newReachability(pk.Hash())
	routing := newRouting(pk, reach)
	c1, c2 := splitChannel(routing.Routes())
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
		ledger,
		&sync.Mutex{},
	}
}

func (r *router) GetAddress() PublicKey {
	return r.pk
}

func (r *router) Connections() []NodeAddress {
	r.lock.Lock()
	defer r.lock.Unlock()
	c := make([]NodeAddress, 0, len(r.connections))
	for k, _ := range r.connections {
		c = append(c, k)
	}
	return c
}

func (r *router) AddConnection(c Connection) {
	r.lock.Lock()
	defer r.lock.Unlock()
	id := c.Key().Hash()
	_, duplicate := r.connections[id]
	if duplicate {
		c.Close()
		return
	}
	r.connections[id] = c

	// Curry the id since the various sub connections don't know about it
	r.routingHandler.AddConnection(id, c)
	r.reachabilityHandler.AddConnection(id, c)
	r.receiptHandler.AddConnection(id, c)
	r.paymentHandler.AddConnection(id, c)
}
