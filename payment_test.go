package node

import (
	"testing"
)

type testPaymentConnection struct {
	in  chan PaymentHash
	out chan PaymentHash
}

func (c testPaymentConnection) SendPayment(p PaymentHash) error {
	c.out <- p
	return nil
}
func (c testPaymentConnection) Payments() <-chan PaymentHash {
	return c.in
}
func (c testPaymentConnection) Close() error {
	close(c.in)
	return nil
}

func makePairedPaymentConnections() (testPaymentConnection, testPaymentConnection) {
	one := make(chan PaymentHash)
	two := make(chan PaymentHash)
	return testPaymentConnection{one, two}, testPaymentConnection{two, one}
}

func TestPaymentHandler(t *testing.T) {
	a1, a2 := NodeAddress("1"), NodeAddress("2")
	p1, p2 := newPayment(a1), newPayment(a2)
	defer p1.Close()
	defer p2.Close()
	c1, c2 := makePairedPaymentConnections()
	defer c1.Close()
	defer c2.Close()
	p1.AddConnection(a2, c1)
	p2.AddConnection(a1, c2)

	go p1.SendPaymentHash(a2, PaymentHash("hash"))

	h := <-p2.PaymentHashes()
	if string(h) != "hash" {
		t.Fatalf("Expected %s == hash", string(h))
	}
}
