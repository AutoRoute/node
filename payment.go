package node

import (
	"log"
	"sync"
)

type PaymentHandler interface {
	AddConnection(NodeAddress, PaymentConnection)
	Receipt(PacketHash)
	AddSentPacket(p Packet, src, next NodeAddress)
	SendPayment(Payment)
	IncomingDebt(NodeAddress) int64
	OutgoingDebt(NodeAddress) int64
}

type paymentImpl struct {
	// debt that other people will pay us
	incoming_debt map[NodeAddress]int64
	// debt that we will pay other people
	outgoing_debt map[NodeAddress]int64
	connections   map[NodeAddress]PaymentConnection
	packetcost    map[PacketHash]int64
	packetnext    map[PacketHash]NodeAddress
	packetsrc     map[PacketHash]NodeAddress
	l             *sync.Mutex
	id            NodeAddress
}

func newPaymentImpl(id NodeAddress) PaymentHandler {
	return &paymentImpl{
		make(map[NodeAddress]int64),
		make(map[NodeAddress]int64),
		make(map[NodeAddress]PaymentConnection),
		make(map[PacketHash]int64),
		make(map[PacketHash]NodeAddress),
		make(map[PacketHash]NodeAddress),
		&sync.Mutex{},
		id}
}

func (p *paymentImpl) AddConnection(id NodeAddress, c PaymentConnection) {
	p.l.Lock()
	defer p.l.Unlock()
	p.connections[id] = c
	go p.handleConnection(c)
}

func (p *paymentImpl) handleConnection(c PaymentConnection) {
	for pay := range c.Payments() {
		if pay.Verify() != nil {
			log.Printf("Error verifying payment: %s", pay.Verify())
			continue
		}
		p.l.Lock()
		p.incoming_debt[pay.Source()] -= pay.Amount()
		p.l.Unlock()
	}
}

func (p *paymentImpl) IncomingDebt(n NodeAddress) int64 {
	p.l.Lock()
	defer p.l.Unlock()
	return p.incoming_debt[n]
}

func (p *paymentImpl) OutgoingDebt(n NodeAddress) int64 {
	p.l.Lock()
	defer p.l.Unlock()
	return p.outgoing_debt[n]
}

func (p *paymentImpl) Receipt(h PacketHash) {
	p.l.Lock()
	defer p.l.Unlock()
	cost, ok := p.packetcost[h]
	if !ok {
		log.Printf("unrecognized hash")
		return
	}
	p.incoming_debt[p.packetsrc[h]] += cost
	p.outgoing_debt[p.packetnext[h]] += cost
}

func (p *paymentImpl) AddSentPacket(pack Packet, src, next NodeAddress) {
	p.l.Lock()
	defer p.l.Unlock()
	p.packetcost[pack.Hash()] = pack.Amount()
	p.packetnext[pack.Hash()] = next
	p.packetsrc[pack.Hash()] = src
}

func (p *paymentImpl) SendPayment(y Payment) {
	p.l.Lock()
	defer p.l.Unlock()
	p.connections[y.Destination()].SendPayment(y)
	p.outgoing_debt[y.Destination()] -= y.Amount()
}