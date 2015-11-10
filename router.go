package node

import (
	"expvar"
	"fmt"
	"sync"
)

var connections_export *expvar.Map
var id_export *expvar.String

func init() {
	connections_export = expvar.NewMap("connections")
	id_export = expvar.NewString("id")
}

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
	*ledger

	lock *sync.Mutex
	quit chan bool
}

func newRouter(pk PublicKey) *router {
	id_export.Set(fmt.Sprintf("%x", pk.Hash()))
	reach := newReachability(pk.Hash())
	routing := newRouting(pk, reach)
	c1, c2, quit := splitChannel(routing.Routes())
	receipt := newReceipt(pk.Hash(), c1)
	ledger := newLedger(pk.Hash(), receipt.PacketHashes(), c2)
	return &router{
		pk,
		make(map[NodeAddress]Connection),
		reach,
		routing,
		receipt,
		ledger,
		&sync.Mutex{},
		quit,
	}
}

func (r *router) GetAddress() PublicKey {
	return r.pk
}

func (r *router) Connections() []Connection {
	r.lock.Lock()
	defer r.lock.Unlock()
	c := make([]Connection, 0, len(r.connections))
	for _, k := range r.connections {
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
	connections_export.Add(fmt.Sprintf("%x", c.Key().Hash()), 1)
}

func (r *router) Close() error {
	r.reachabilityHandler.Close()
	r.routingHandler.Close()
	r.receiptHandler.Close()
	r.ledger.Close()
	close(r.quit)
	return nil
}
