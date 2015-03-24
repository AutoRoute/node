package node

import (
	"testing"
)

type testPaymentConnection struct {
	in  chan Payment
	out chan Payment
}

func (c testPaymentConnection) SendPayment(p Payment) error {
	c.out <- p
	return nil
}
func (c testPaymentConnection) Payments() <-chan Payment {
	return c.in
}

func makePairedPaymentConnections() (PaymentConnection, PaymentConnection) {
	one := make(chan Payment)
	two := make(chan Payment)
	return testPaymentConnection{one, two}, testPaymentConnection{two, one}
}

func TestPaymentHandler(t *testing.T) {
	c1, c2 := makePairedPaymentConnections()
	a1 := NodeAddress("1")
	a2 := NodeAddress("2")
	p1 := newPaymentImpl(a1)
	p2 := newPaymentImpl(a2)
	p1.AddConnection(a2, c1)
	p2.AddConnection(a1, c2)

	t1 := testPacket(a2)

	p1.AddSentPacket(t1, a1, a2)

}
