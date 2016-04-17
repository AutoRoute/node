package internal

import (
	"expvar"
	"fmt"
	"sync"

	"github.com/AutoRoute/node/types"
)

var connections_export *expvar.Map
var id_export *expvar.String

func init() {
	connections_export = expvar.NewMap("connections")
	id_export = expvar.NewString("id")
}

// A Router handles all routing tasks that don't involve the local machine
// including connection management, reachability handling, packet receipt
// relaying, and (outstanding) payment tracking. AKA anything which doesn't
// need the private key.
type Router struct {
	pk PublicKey
	// A map of public key hashes to connections
	connections map[types.NodeAddress]Connection

	*routingHandler
	*reachabilityHandler
	*receiptHandler
	*Ledger

	lock *sync.Mutex
	quit chan bool
}

func NewRouter(pk PublicKey) *Router {
	id_export.Set(fmt.Sprintf("%x", pk.Hash()))
	reach := newReachability(pk.Hash())
	// TODO(daniel): Eventually, add flags that allow a user to specify alternate
	// routing algorithms.
	algorithm := newBandwidthRouting(reach)
	routing := newRoutingHandler(pk, algorithm)
	c1, c2, quit := splitChannel(routing.Routes())
	receipt := newReceipt(pk.Hash(), c1)
	Ledger := newLedger(pk.Hash(), receipt.PacketHashes(), c2)
	return &Router{
		pk,
		make(map[types.NodeAddress]Connection),
		routing,
		reach,
		receipt,
		Ledger,
		&sync.Mutex{},
		quit,
	}
}

func (r *Router) GetAddress() PublicKey {
	return r.pk
}

func (r *Router) Connections() []Connection {
	r.lock.Lock()
	defer r.lock.Unlock()
	c := make([]Connection, 0, len(r.connections))
	for _, k := range r.connections {
		c = append(c, k)
	}
	return c
}

func (r *Router) AddConnection(c Connection) {
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

func (r *Router) Close() error {
	r.reachabilityHandler.Close()
	r.routingHandler.Close()
	r.receiptHandler.Close()
	r.Ledger.Close()
	close(r.quit)
	return nil
}
