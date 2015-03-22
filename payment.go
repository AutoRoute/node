package node

import (
	"log"
	"sync"
)

type PaymentHandler interface {
	AddConnection(NodeAddress, PaymentConnection)
	Receipt(PacketHash)
	AddSentPacket(p Packet, src, next NodeAddress)
}

type paymentImpl struct {
    outstanding_debt map[NodeAddress]int64
	connections map[NodeAddress]PaymentConnection
    packetcost map[PacketHash]int64
    packetnext map[PacketHash]NodeAddress
	l           *sync.Mutex
	id          NodeAddress
}

func (p paymentImpl) AddConnection(id NodeAddress, c PaymentConnection) {
    p.l.Lock()
    defer p.l.Unlock()
    p.connections[id] = c
    go p.handleConnection(c)
}

func (p paymentImpl) handleConnection(c PaymentConnection) {
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

func (p paymentImpl) Receipt(h PacketHash) {
    p.l.Lock()
    defer p.l.Unlock()
    cost, ok := p.packetcost[h]
    if !ok {
        return
    }
    p.outstanding_debt[p.packetnext[h]] += cost
}

func (p paymentImpl) AddSentPacket(pack Packet, src, next NodeAddress) {
    p.l.Lock()
    defer p.l.Unlock()
    p.packetcost[pack.Hash()] = pack.Amount()
    p.packetnext[pack.Hash()] = next
}
