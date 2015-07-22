package node

import (
	"log"
	"sync"
	"time"
)

// This keeps track of our outstanding owed payments and provides an interface
// to send payments. It does not create payments on its own.
type Ledger interface {
	RecordPayment(Payment)
	// There are really only a few interesting pieces of data you will want to
	// know about. How much we should pay someone, and how much debt + how long
	// it has been since someone paid us (aka stop relaying to them if they haven't
	// paid recently enough or the outstanding balance is too high).
	IncomingDebt(NodeAddress) (int64, time.Time)
	OutgoingDebt(NodeAddress) (int64, time.Time)
}

// Debt is recorded as amount + time. The assumption is that debt is payed
// oldest first.
type debt struct {
	time   time.Time
	amount int64
}

func payDebt(debts []debt, amount int64) []debt {
	for _, d := range debts {
		if d.amount < amount {
			// not needed since removing
			amount -= d.amount
			d.amount = 0
			debts = debts[1:]
			continue
		}
		d.amount -= amount
		amount = 0
		debts[0] = d
		break
	}
	if amount != 0 {
		// Overpayment?
		log.Print("Overpayment?: ", amount)
	}
	return debts
}

func sumDebt(debts []debt) (int64, time.Time) {
	var s time.Time
	a := int64(0)
	if len(debts) > 0 {
		s = debts[0].time
	}
	for _, i := range debts {
		a += i.amount
	}
	return a, s
}

type ledger struct {
	// debt that other people will pay us
	incoming_debt map[NodeAddress][]debt
	// debt that we will pay other people
	outgoing_debt map[NodeAddress][]debt
	packets       map[PacketHash]routingDecision
	l             *sync.Mutex
	id            NodeAddress
}

func newLedger(id NodeAddress, c <-chan PacketHash, d <-chan routingDecision) Ledger {
	p := &ledger{
		make(map[NodeAddress][]debt),
		make(map[NodeAddress][]debt),
		make(map[PacketHash]routingDecision),
		&sync.Mutex{},
		id}
	go p.handleReceipt(c)
	go p.sentPackets(d)
	return p
}

func (p *ledger) IncomingDebt(n NodeAddress) (int64, time.Time) {
	p.l.Lock()
	defer p.l.Unlock()
	return sumDebt(p.incoming_debt[n])
}

func (p *ledger) OutgoingDebt(n NodeAddress) (int64, time.Time) {
	p.l.Lock()
	defer p.l.Unlock()
	return sumDebt(p.outgoing_debt[n])
}

func (p *ledger) handleReceipt(c <-chan PacketHash) {
	for h := range c {
		p.l.Lock()
		i, ok := p.packets[h]
		if !ok {
			log.Printf("unrecognized hash")
			p.l.Unlock()
			continue
		}
		d := debt{time.Now(), i.amount}
		p.incoming_debt[i.source] = append(p.incoming_debt[i.source], d)
		p.outgoing_debt[i.nexthop] = append(p.outgoing_debt[i.nexthop], d)
		p.l.Unlock()
	}
}

func (p *ledger) sentPackets(c <-chan routingDecision) {
	for d := range c {
		p.l.Lock()
		p.packets[d.hash] = d
		p.l.Unlock()
	}
}

func (p *ledger) RecordPayment(y Payment) {
	p.l.Lock()
	defer p.l.Unlock()
	p.incoming_debt[y.Source] = payDebt(p.incoming_debt[y.Source], y.Amount)
	p.outgoing_debt[y.Destination] = payDebt(p.outgoing_debt[y.Destination], y.Amount)
}
