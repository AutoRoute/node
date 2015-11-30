package node

import (
	"log"
	"sync"
)

// This keeps track of our outstanding owed payments and provides an interface
// to send payments. It does not create payments on its own.
// There are really only a few interesting pieces of data you will want to
// know about. How much we should pay someone, and how much debt + how long
// it has been since someone paid us (aka stop relaying to them if they haven't
// paid recently enough or the outstanding balance is too high).
type ledger struct {
	// debt that other people will pay us
	incoming_debt map[NodeAddress]int64
	// debt that we will pay other people
	outgoing_debt    map[NodeAddress]int64
	packets          map[PacketHash]routingDecision
	payment_channels map[string]chan uint64
	l                *sync.Mutex
	id               NodeAddress
	quit             chan bool
}

func newLedger(id NodeAddress, c <-chan PacketHash, d <-chan routingDecision) *ledger {
	p := &ledger{
		make(map[NodeAddress]int64),
		make(map[NodeAddress]int64),
		make(map[PacketHash]routingDecision),
		make(map[string]chan uint64),
		&sync.Mutex{},
		id,
		make(chan bool),
	}
	go p.handleReceipt(c)
	go p.sentPackets(d)
	return p
}

func (p *ledger) IncomingDebt(n NodeAddress) int64 {
	p.l.Lock()
	defer p.l.Unlock()
	return p.incoming_debt[n]
}

func (p *ledger) OutgoingDebt(n NodeAddress) int64 {
	p.l.Lock()
	defer p.l.Unlock()
	return p.outgoing_debt[n]
}

func (p *ledger) AddAddress(address string, c chan uint64) {
	p.l.Lock()
	defer p.l.Unlock()
	p.payment_channels[address] = c
}

func (p *ledger) AddConnection(n NodeAddress, c Connection) {
	go p.handlePayments(n, c)
}

func (p *ledger) handleReceipt(c <-chan PacketHash) {
	for {
		select {
		case h := <-c:
			p.l.Lock()
			i, ok := p.packets[h]
			if !ok {
				log.Printf("unrecognized hash")
				p.l.Unlock()
				continue
			}
			p.incoming_debt[i.source] += i.amount
			p.outgoing_debt[i.nexthop] += i.amount
			p.l.Unlock()
		case <-p.quit:
			return
		}
	}
}

func (p *ledger) sentPackets(c <-chan routingDecision) {
	for {
		select {
		case d := <-c:
			p.l.Lock()
			p.packets[d.hash] = d
			p.l.Unlock()
		case <-p.quit:
			return
		}
	}
}

func (p *ledger) handlePayments(n NodeAddress, c Connection) {
	p.l.Lock()
	ch := p.payment_channels[c.MetaData().Payment_Address]
	p.l.Unlock()
	for {
		select {
		case amount := <-ch:
			log.Printf("Received payment of %d to %q", amount, c.MetaData().Payment_Address)
			p.l.Lock()
			p.incoming_debt[n] -= int64(amount)
			p.l.Unlock()
		case <-p.quit:
			return
		}
	}
}

// Waits for the payment to be confirmed and records it in the ledger.
func (p *ledger) RecordPayment(destination NodeAddress, amount int64, confirmed chan bool) {
	ok := <-confirmed
	if ok {
		p.l.Lock()
		p.incoming_debt[p.id] -= amount
		p.outgoing_debt[destination] -= amount
		p.l.Unlock()
	}
}

func (p *ledger) Close() error {
	close(p.quit)
	return nil
}
