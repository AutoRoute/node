package internal

import (
	"expvar"
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/AutoRoute/node/types"
)

var packets_sent *expvar.Map
var packets_received *expvar.Map
var packets_dropped *expvar.Int

func init() {
	packets_sent = expvar.NewMap("packets_sent")
	packets_received = expvar.NewMap("packets_received")
	packets_dropped = expvar.NewInt("packets_dropped")
}

// Represents a permanent record of a routing decision.
type routingDecision struct {
	hash        types.PacketHash
	amount      int64
	source      types.NodeAddress
	destination types.NodeAddress
	nexthop     types.NodeAddress
}

func newRoutingDecision(p types.Packet, src types.NodeAddress, nexthop types.NodeAddress) routingDecision {
	return routingDecision{p.Hash(), p.Amount(), src, p.Destination(), nexthop}
}

// A routing handler takes care of relaying packets and produces notifications
// about it's routing decisions.
type routingHandler struct {
	pk PublicKey
	// A chan down which we send packets destined for ourselves.
	incoming chan types.Packet
	routes   chan routingDecision
	// A map of public key hashes to connections
	connections  map[types.NodeAddress]DataConnection
	reachability *reachabilityHandler
	quit         chan bool
	// Bandwidth estimator for all nodes.
	bandwidths *bandwidthEstimator
	// Mutex for serializing bandwidthEstimator operations.
	bandwidth_mutex sync.RWMutex
}

func newRouting(pk PublicKey, r *reachabilityHandler) *routingHandler {
	return &routingHandler{
		pk,
		make(chan types.Packet),
		make(chan routingDecision),
		make(map[types.NodeAddress]DataConnection),
		r,
		make(chan bool),
		NewBandwidthEstimator(),
		sync.RWMutex{},
	}
}

func (r *routingHandler) AddConnection(id types.NodeAddress, c DataConnection) {
	r.connections[id] = c
	go r.handleData(id, c)
}

func (r *routingHandler) handleData(id types.NodeAddress, p DataConnection) {
	for {
		select {
		case packet, ok := <-p.Packets():
			if !ok {
				log.Printf("Packet channel closed, exiting")
				return
			}
			err := r.sendPacket(packet, id)
			if err != nil {
				log.Printf("%x: Dropping packet destined to %x: %v", r.pk.Hash(), packet.Destination(), err)
			}
		case <-r.quit:
			return
		}
	}
}

func (r *routingHandler) SendPacket(p types.Packet) error {
	return r.sendPacket(p, r.pk.Hash())
}

func (r *routingHandler) sendPacket(p types.Packet, src types.NodeAddress) error {
	packets_received.Add(fmt.Sprintf("%x", src), 1)
	if p.Destination() == r.pk.Hash() {
		packets_sent.Add(fmt.Sprintf("%x", r.pk.Hash()), 1)
		r.incoming <- p
		go r.notifyDecision(p, src, r.pk.Hash())
		return nil
	}
	possible_next, err := r.reachability.FindNextHop(p.Destination(), src)
	if err != nil {
		packets_dropped.Add(1)
		return err
	}

	// Decide which one of our possible destinations to send it to.
	r.bandwidth_mutex.RLock()
	weights := r.bandwidths.GetWeights(possible_next)
	r.bandwidth_mutex.RUnlock()
	// Choose a random destination given our weights.
	choice := rand.Float64()
	// We pick the last one because if it should be the last one, floating-point
	// weirdness could cause us not to pick it.
	next := possible_next[len(possible_next)-1]
	total := 0.0
	for i, weight := range weights {
		total += weight
		log.Printf("Total: %f\n", total)
		log.Printf("Choice: %f\n", choice)
		if total >= choice {
			next = possible_next[i]
		}
	}

	packets_sent.Add(fmt.Sprintf("%x", next), 1)
	go r.notifyDecision(p, src, next)

	r.bandwidth_mutex.Lock()
	r.bandwidths.WillSendPacket(next)
	r.bandwidth_mutex.Unlock()
	err = r.connections[next].SendPacket(p)
	if err != nil {
		log.Print("Error sending packet.\n")
		return err
	}
	r.bandwidth_mutex.Lock()
	r.bandwidths.SentPacket(next, int64(len(p.Data)))
	r.bandwidth_mutex.Unlock()
	return nil
}

func (r *routingHandler) notifyDecision(p types.Packet, src, next types.NodeAddress) {
	r.routes <- routingDecision{p.Hash(), p.Amount(), src, p.Destination(), next}
}

func (r *routingHandler) Routes() <-chan routingDecision {
	return r.routes
}

func (r *routingHandler) Packets() <-chan types.Packet {
	return r.incoming
}

func (r *routingHandler) Close() error {
	close(r.quit)
	return nil
}
