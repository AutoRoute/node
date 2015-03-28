package node

import (
	"log"
)

// The brains of everything, takes in connections and wires them togethor.
type Router interface {
	AddConnection(Connection)
	DataConnection
	GetAddress() PublicKey
}

type routerImpl struct {
	pk       PublicKey
	incoming chan Packet
	// A struct which maintains reachability information
	reachability MapHandler
	// A map of public key hashes to connections
	connections map[NodeAddress]Connection
	payments    ReceiptHandler
}

type NullAction struct{}

func (n NullAction) Receipt(PacketHash) {}

func newRouterImpl(pk PublicKey) Router {
	return &routerImpl{pk, make(chan Packet), newMapImpl(pk.Hash()), make(map[NodeAddress]Connection), newReceiptImpl(pk.Hash(), NullAction{})}
}

func (r *routerImpl) GetAddress() PublicKey {
	return r.pk
}

func (r *routerImpl) AddConnection(c Connection) {
	id := c.Key().Hash()
	_, duplicate := r.connections[id]
	if duplicate {
		c.Close()
		return
	}
	r.connections[id] = c

	// Curry the id since the various sub connections don't know about it
	r.reachability.AddConnection(id, c)
	go r.handleData(id, c)
	//go r.handleReceipts(id, c)
}

func (r *routerImpl) handleData(id NodeAddress, p DataConnection) {
	for packet := range p.Packets() {
		err := r.SendPacket(packet)
		if err != nil {
			log.Printf("%q: Dropping packet destined to %q: %q", r.pk.Hash(), packet.Destination(), err)
			continue
		}
	}
}

func (r *routerImpl) SendPacket(p Packet) error {
	if p.Destination() == r.pk.Hash() {
		log.Printf("%q: Routing packet to self", r.pk.Hash())
		r.incoming <- p
		return nil
	}
	rid, err := r.reachability.FindConnection(p.Destination())
	if err != nil {
		return err
	}
	log.Printf("%q: Routing packet to %q", r.pk.Hash(), rid)
	return r.connections[rid].SendPacket(p)
}

func (r *routerImpl) Packets() <-chan Packet {
	return r.incoming
}

//func (r *routerImpl) handleReceipts(id NodeAddress, p ReceiptConnection) {}
