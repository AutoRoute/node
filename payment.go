package node

import (
	"log"
	"sync"
)

type PaymentHandler interface {
	AddConnection(NodeAddress, PaymentConnection)
	Receipt(PacketHash)
	AddSentPacket(p Packet, src, next NodeAddress)
	OutstandingDebt(NodeAddress) int64
}

type paymentImpl struct {
	outstanding_debt map[NodeAddress]int64
	connections      map[NodeAddress]PaymentConnection
	packetcost       map[PacketHash]int64
	packetnext       map[PacketHash]NodeAddress
	l                *sync.Mutex
	id               NodeAddress
}

func newPaymentImpl(id NodeAddress) PaymentHandler {
	return &paymentImpl{make(map[NodeAddress]int64),
		make(map[NodeAddress]PaymentConnection),
		make(map[PacketHash]int64),
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
		p.outstanding_debt[pay.Source()] -= pay.Amount()
		p.l.Unlock()
	}
}

func (p *paymentImpl) OutstandingDebt(n NodeAddress) int64 {
	p.l.Lock()
	defer p.l.Unlock()
	return p.outstanding_debt[n]
}

func (p *paymentImpl) Receipt(h PacketHash) {
	p.l.Lock()
	defer p.l.Unlock()
	cost, ok := p.packetcost[h]
	if !ok {
		return
	}
	p.outstanding_debt[p.packetnext[h]] += cost
}

func (p *paymentImpl) AddSentPacket(pack Packet, src, next NodeAddress) {
	p.l.Lock()
	defer p.l.Unlock()
	p.packetcost[pack.Hash()] = pack.Amount()
	p.packetnext[pack.Hash()] = next
}
