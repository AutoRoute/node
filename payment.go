package node

import (
	"sync"
)

// This keeps track of our outstanding owed payments and provides an interface
// to send payments. It does not create payments on its own.
type PaymentHandler interface {
	AddConnection(NodeAddress, PaymentConnection)
	SendPaymentHash(NodeAddress, PaymentHash) error
	PaymentHashes() <-chan PaymentHash
}

type payment struct {
	connections map[NodeAddress]PaymentConnection
	c           chan PaymentHash
	l           *sync.Mutex
	id          NodeAddress
}

func newPayment(id NodeAddress) PaymentHandler {
	p := &payment{
		make(map[NodeAddress]PaymentConnection),
		make(chan PaymentHash),
		&sync.Mutex{},
		id}
	return p
}

func (p *payment) PaymentHashes() <-chan PaymentHash {
	return p.c
}

func (p *payment) AddConnection(id NodeAddress, c PaymentConnection) {
	p.l.Lock()
	defer p.l.Unlock()
	p.connections[id] = c
	go p.handleConnection(c)
}

func (p *payment) handleConnection(c PaymentConnection) {
	for hash := range c.Payments() {
		p.c <- hash
	}
}

func (p *payment) SendPaymentHash(id NodeAddress, y PaymentHash) error {
	p.l.Lock()
	defer p.l.Unlock()
	return p.connections[id].SendPayment(y)
}
