package node

import (
	"log"
)

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
}

func newRouting(pk PublicKey, r *reachabilityHandler) *routingHandler {
	return &routingHandler{
		pk,
		make(chan Packet),
		make(chan routingDecision),
		make(map[NodeAddress]DataConnection),
		r}
}

func (r *routingHandler) AddConnection(id NodeAddress, c DataConnection) {
	r.connections[id] = c
	go r.handleData(id, c)
}

func (r *routingHandler) handleData(id NodeAddress, p DataConnection) {
	for packet := range p.Packets() {
		err := r.sendPacket(packet, id)
		if err != nil {
			log.Printf("%q: Dropping packet destined to %q: %q", r.pk.Hash(), packet.Destination(), err)
			continue
		}
	}
}

func (r *routingHandler) SendPacket(p Packet) error {
	return r.sendPacket(p, r.pk.Hash())
}

func (r *routingHandler) sendPacket(p Packet, src NodeAddress) error {
	if p.Destination() == r.pk.Hash() {
		r.incoming <- p
		r.notifyDecision(p, src, r.pk.Hash())
		return nil
	}
	next, err := r.reachability.FindNextHop(p.Destination())
	if err != nil {
		return err
	}
	r.notifyDecision(p, src, next)
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
	close(r.incoming)
	close(r.routes)
	return nil
}
