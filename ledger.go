package node

import (
	"log"
	"sync"
	"time"
)

// Debt is recorded as amount + time. The assumption is that debt is payed
// oldest first.
type debt struct {
	time   time.Time
	amount int64
}

// Pays debt, removing the oldest first.
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

// Sums the debt and returns the earliest time.
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

// This keeps track of our outstanding owed payments and provides an interface
// to send payments. It does not create payments on its own.
// There are really only a few interesting pieces of data you will want to
// know about. How much we should pay someone, and how much debt + how long
// it has been since someone paid us (aka stop relaying to them if they haven't
// paid recently enough or the outstanding balance is too high).
type ledger struct {
	// debt that other people will pay us
	incoming_debt map[NodeAddress][]debt
	// debt that we will pay other people
	outgoing_debt    map[NodeAddress][]debt
	packets          map[PacketHash]routingDecision
	payment_channels map[string]chan uint64
	l                *sync.Mutex
	id               NodeAddress
	quit             chan bool
}

func newLedger(id NodeAddress, c <-chan PacketHash, d <-chan routingDecision) *ledger {
	p := &ledger{
		make(map[NodeAddress][]debt),
		make(map[NodeAddress][]debt),
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
			d := debt{time.Now(), i.amount}
			p.incoming_debt[i.source] = append(p.incoming_debt[i.source], d)
			p.outgoing_debt[i.nexthop] = append(p.outgoing_debt[i.nexthop], d)
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
			p.incoming_debt[n] = payDebt(p.incoming_debt[n], int64(amount))
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
		p.incoming_debt[p.id] = payDebt(p.incoming_debt[p.id], amount)
		p.outgoing_debt[destination] = payDebt(p.outgoing_debt[destination], amount)
		p.l.Unlock()
	}
}

func (p *ledger) Close() error {
	close(p.quit)
	return nil
}
