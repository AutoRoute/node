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

type testPayment struct {
	src NodeAddress
	dst NodeAddress
	amt int64
}

func (t testPayment) Source() NodeAddress      { return t.src }
func (t testPayment) Destination() NodeAddress { return t.dst }
func (t testPayment) Verify() error            { return nil }
func (t testPayment) Amount() int64            { return t.amt }

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
	if p1.IncomingDebt(a1) != 0 {
		t.Fatalf("Expected a1 incomingdebt %d got %d", 0, p1.IncomingDebt(a1))
	}
	if p1.OutgoingDebt(a2) != 0 {
		t.Fatalf("Expected a2 outgoingdebt %d got %d", 0, p1.OutgoingDebt(a2))
	}
	p1.Receipt(t1.Hash())
	if p1.IncomingDebt(a1) != 1 {
		t.Fatalf("Expected a1 incomingdebt %d got %d", 1, p1.IncomingDebt(a1))
	}
	if p1.OutgoingDebt(a2) != 1 {
		t.Fatalf("Expected a2 outgoingdebt %d got %d", 1, p1.OutgoingDebt(a2))
	}

	p2.AddSentPacket(t1, a1, a2)
	p2.Receipt(t1.Hash())
	if p2.IncomingDebt(a1) != 1 {
		t.Fatalf("Expected a1 incomingdebt %d got %d", 1, p2.IncomingDebt(a1))
	}

	payment := testPayment{a1, a2, 1}
	p1.SendPayment(payment)
	if p2.IncomingDebt(a1) != 0 {
		t.Fatalf("Expected a1 incomingdebt %d got %d", 0, p2.IncomingDebt(a1))
	}
	if p1.OutgoingDebt(a2) != 0 {
		t.Fatalf("Expected a2 outgoingdebt %d got %d", 0, p1.OutgoingDebt(a2))
	}
}
