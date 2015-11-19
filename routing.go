package node

import (
	"expvar"
	"fmt"
	"log"
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
	hash        PacketHash
	amount      int64
	source      NodeAddress
	destination NodeAddress
	nexthop     NodeAddress
}

func newRoutingDecision(p Packet, src NodeAddress, nexthop NodeAddress) routingDecision {
	return routingDecision{p.Hash(), p.Amount(), src, p.Destination(), nexthop}
}

// A routing handler takes care of relaying packets and produces notifications
// about it's routing decisions.
type routingHandler struct {
	pk PublicKey
	// A chan down which we send packets destined for ourselves.
	incoming chan Packet
	routes   chan routingDecision
	// A map of public key hashes to connections
	connections  map[NodeAddress]DataConnection
	reachability *reachabilityHandler
	quit         chan bool
}

func newRouting(pk PublicKey, r *reachabilityHandler) *routingHandler {
	return &routingHandler{
		pk,
		make(chan Packet),
		make(chan routingDecision),
		make(map[NodeAddress]DataConnection),
		r,
		make(chan bool),
	}
}

func (r *routingHandler) AddConnection(id NodeAddress, c DataConnection) {
	r.connections[id] = c
	go r.handleData(id, c)
}

func (r *routingHandler) handleData(id NodeAddress, p DataConnection) {
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

func (r *routingHandler) SendPacket(p Packet) error {
	return r.sendPacket(p, r.pk.Hash())
}

func (r *routingHandler) sendPacket(p Packet, src NodeAddress) error {
	packets_received.Add(fmt.Sprint("%x", src), 1)
	if p.Destination() == r.pk.Hash() {
		packets_sent.Add(fmt.Sprint("%x", r.pk.Hash), 1)
		r.incoming <- p
		go r.notifyDecision(p, src, r.pk.Hash())
		return nil
	}
	next, err := r.reachability.FindNextHop(p.Destination())
	if err != nil {
		packets_dropped.Add(1)
		return err
	}
	packets_sent.Add(fmt.Sprint("%x", next), 1)
	go r.notifyDecision(p, src, next)
	return r.connections[next].SendPacket(p)
}

func (r *routingHandler) notifyDecision(p Packet, src, next NodeAddress) {
	r.routes <- routingDecision{p.Hash(), p.Amount(), src, p.Destination(), next}
}

func (r *routingHandler) Routes() <-chan routingDecision {
	return r.routes
}

func (r *routingHandler) Packets() <-chan Packet {
	return r.incoming
}

func (r *routingHandler) Close() error {
	close(r.quit)
	return nil
}
