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

func makePairedPaymentConnections() (PaymentConnection, PaymentConnection) {
	one := make(chan PaymentHash)
	two := make(chan PaymentHash)
	return testPaymentConnection{one, two}, testPaymentConnection{two, one}
}

func TestPaymentHandler(t *testing.T) {
	c1, c2 := makePairedPaymentConnections()
	a1, a2 := NodeAddress("1"), NodeAddress("2")
	p1, p2 := newPayment(a1), newPayment(a2)
	p1.AddConnection(a2, c1)
	p2.AddConnection(a1, c2)

	go p1.SendPaymentHash(a2, PaymentHash("hash"))

	h := <-p2.PaymentHashes()
	if string(h) != "hash" {
		t.Fatalf("Expected %s == hash", string(h))
	}

}
