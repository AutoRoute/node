package node

import (
	"log"
)

// A router handles all routing tasks that don't involve the local machine
// including connection management, reachability handling, packet receipt
// relaying, and (outstanding) payment tracking.
type Router interface {
	AddConnection(Connection)
	DataConnection
	GetAddress() PublicKey
	SendReceipt(PacketReceipt)
	SendPayment(Payment)
}

type routerImpl struct {
	pk PublicKey
	// A chan down which we send packets destined for ourselves.
	incoming chan Packet
	// A map of public key hashes to connections
	connections map[NodeAddress]Connection

	reachability MapHandler
	receipt      ReceiptHandler
	payment      PaymentHandler
}

func newRouterImpl(pk PublicKey) Router {
	payment := newPaymentImpl(pk.Hash())
	return &routerImpl{
		pk,
		make(chan Packet),
		make(map[NodeAddress]Connection),
		newMapImpl(pk.Hash()),
		newReceiptImpl(pk.Hash(), payment),
		payment}
}

func (r *routerImpl) GetAddress() PublicKey {
	return r.pk
}

func (r *routerImpl) SendReceipt(p PacketReceipt) {
	r.receipt.SendReceipt(p)
}

func (r *routerImpl) SendPayment(p Payment) {
	r.payment.SendPayment(p)
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
	r.receipt.AddConnection(id, c)
	r.payment.AddConnection(id, c)
	go r.handleData(id, c)
}

func (r *routerImpl) handleData(id NodeAddress, p DataConnection) {
	for packet := range p.Packets() {
		err := r.sendPacket(packet, id)
		if err != nil {
			log.Printf("%q: Dropping packet destined to %q: %q", r.pk.Hash(), packet.Destination(), err)
			continue
		}
	}
}

func (r *routerImpl) SendPacket(p Packet) error {
	return r.sendPacket(p, r.pk.Hash())
}

func (r *routerImpl) sendPacket(p Packet, src NodeAddress) error {
	if p.Destination() == r.pk.Hash() {
		log.Printf("%q: Routing packet to self", r.pk.Hash())
		r.incoming <- p
		return nil
	}
	next, err := r.reachability.FindConnection(p.Destination())
	if err != nil {
		return err
	}
	log.Printf("%q: Routing packet to %q", r.pk.Hash(), next)
	go r.receipt.AddSentPacket(p, src, next)
	go r.payment.AddSentPacket(p, src, next)
	return r.connections[next].SendPacket(p)
}

func (r *routerImpl) Packets() <-chan Packet {
	return r.incoming
}
