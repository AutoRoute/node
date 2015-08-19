package node

import (
	"sync"
)

// This keeps track of our outstanding owed payments and provides an interface
// to send payments. It does not create payments on its own.
type paymentHandler struct {
	connections map[NodeAddress]PaymentConnection
	c           chan PaymentHash
	l           *sync.Mutex
	id          NodeAddress
	quit        chan bool
}

func newPayment(id NodeAddress) *paymentHandler {
	p := &paymentHandler{
		make(map[NodeAddress]PaymentConnection),
		make(chan PaymentHash),
		&sync.Mutex{},
		id,
		make(chan bool),
	}
	return p
}

func (p *paymentHandler) PaymentHashes() <-chan PaymentHash {
	return p.c
}

func (p *paymentHandler) AddConnection(id NodeAddress, c PaymentConnection) {
	p.l.Lock()
	defer p.l.Unlock()
	p.connections[id] = c
	go p.handleConnection(c)
}

func (p *paymentHandler) handleConnection(c PaymentConnection) {
	for {
		select {
		case hash, ok := <-c.Payments():
			if !ok {
				return
			}
			p.c <- hash
		case <-p.quit:
			return
		}
	}
}

func (p *paymentHandler) SendPaymentHash(id NodeAddress, y PaymentHash) error {
	p.l.Lock()
	defer p.l.Unlock()
	return p.connections[id].SendPayment(y)
}

func (p *paymentHandler) Close() error {
	close(p.quit)
	return nil
}
