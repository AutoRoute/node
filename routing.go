package node

import (
	"log"
)

// Represents a permanent record of a routing decision.
type RoutingDecision struct {
	hash        PacketHash
	amount      int64
	source      NodeAddress
	destination NodeAddress
	nexthop     NodeAddress
}

func NewRoutingDecision(p Packet, src NodeAddress, nexthop NodeAddress) RoutingDecision {
	return RoutingDecision{p.Hash(), p.Amount(), src, p.Destination(), nexthop}
}

// A routing handler takes care of relaying packets and produces notifications
// about it's routing decisions.
type RoutingHandler interface {
	AddConnection(NodeAddress, DataConnection)
	// For thing which are routed to "us"
	DataConnection
	// The actions we've taken
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
		r.incoming <- p
		go r.notifyDecision(p, src, r.pk.Hash())
		return nil
	}
	next, err := r.reachability.FindNextHop(p.Destination())
	if err != nil {
		return err
	}
	go r.notifyDecision(p, src, next)
	return r.connections[next].SendPacket(p)
}

func (r *routing) notifyDecision(p Packet, src, next NodeAddress) {
	r.routes <- RoutingDecision{p.Hash(), p.Amount(), src, p.Destination(), next}
}

func (r *routing) Routes() <-chan RoutingDecision {
	return r.routes
}

func (r *routing) Packets() <-chan Packet {
	return r.incoming
}
