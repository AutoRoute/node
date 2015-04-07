package node

import (
	"log"
)

type RoutingDecision struct {
	p       Packet
	nexthop NodeAddress
	source  NodeAddress
}

// A routing handler takes care of relaying packets and produces notifications
// about it's routing decisions.
type RoutingHandler interface {
	AddConnection(NodeAddress, DataConnection)
	DataConnection
	Routes() <-chan RoutingDecision
}

type routing struct {
	pk PublicKey
	// A chan down which we send packets destined for ourselves.
	incoming chan Packet
	routes   chan RoutingDecision
	// A map of public key hashes to connections
	connections  map[NodeAddress]DataConnection
	reachability ReachabilityHandler
}

func newRouting(pk PublicKey, r ReachabilityHandler) RoutingHandler {
	return &routing{
		pk,
		make(chan Packet),
		make(chan RoutingDecision),
		make(map[NodeAddress]DataConnection),
		r}
}

func (r *routing) AddConnection(id NodeAddress, c DataConnection) {
	r.connections[id] = c
	go r.handleData(id, c)
}

func (r *routing) handleData(id NodeAddress, p DataConnection) {
	for packet := range p.Packets() {
		err := r.sendPacket(packet, id)
		if err != nil {
			log.Printf("%q: Dropping packet destined to %q: %q", r.pk.Hash(), packet.Destination(), err)
			continue
		}
	}
}

func (r *routing) SendPacket(p Packet) error {
	return r.sendPacket(p, r.pk.Hash())
}

func (r *routing) sendPacket(p Packet, src NodeAddress) error {
	if p.Destination() == r.pk.Hash() {
		log.Printf("%q: Routing packet to self", r.pk.Hash())
		r.incoming <- p
		return nil
	}
	next, err := r.reachability.FindNextHop(p.Destination())
	if err != nil {
		return err
	}
	log.Printf("%q: Routing packet to %q", r.pk.Hash(), next)
	go r.notifyDecision(p, src, next)
	return r.connections[next].SendPacket(p)
}

func (r *routing) notifyDecision(p Packet, src, next NodeAddress) {
	r.routes <- RoutingDecision{p, src, next}
}

func (r *routing) Routes() <-chan RoutingDecision {
	return r.routes
}

func (r *routing) Packets() <-chan Packet {
	return r.incoming
}
