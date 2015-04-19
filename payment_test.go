package node

import (
	"testing"
	"time"
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

func WaitForIncomingDebt(t *testing.T, p PaymentHandler, a NodeAddress, m int64) {
	timeout := time.After(time.Second)
	tick := time.Tick(time.Millisecond * 5)
	for {
		select {
		case _ = <-timeout:
			t.Fatalf("Timeout witing for IncomingDebt(%v) from %v in %v", m, a, p)
			return
		case _ = <-tick:
			if d, _ := p.IncomingDebt(a); d == m {
				return
			}
		}
	}
}

func WaitForOutgoingDebt(t *testing.T, p PaymentHandler, a NodeAddress, m int64) {
	timeout := time.After(time.Second)
	tick := time.Tick(time.Millisecond * 5)
	for {
		select {
		case _ = <-timeout:
			t.Fatalf("Timeout witing for OutgoingDebt(%v) from %v in %v", m, a, p)
			return
		case _ = <-tick:
			if d, _ := p.OutgoingDebt(a); d == m {
				return
			}
		}
	}
}

func TestPaymentHandler(t *testing.T) {
	h1, h2 := make(chan PacketHash), make(chan PacketHash)
	c1, c2 := makePairedPaymentConnections()
	a1, a2 := NodeAddress("1"), NodeAddress("2")
	i1, i2 := make(chan RoutingDecision), make(chan RoutingDecision)
	p1, p2 := newPayment(a1, h1, i1), newPayment(a2, h2, i2)
	p1.AddConnection(a2, c1)
	p2.AddConnection(a1, c2)

	t1 := testPacket(a2)
	t2 := testPacket(NodeAddress("3"))
	i1 <- NewRoutingDecision(t1, a1, a2)
	WaitForIncomingDebt(t, p1, a1, 0)
	WaitForOutgoingDebt(t, p1, a2, 0)
	h1 <- t1.Hash()
	owed := t1.Amount()
	WaitForIncomingDebt(t, p1, a1, owed)
	WaitForOutgoingDebt(t, p1, a2, owed)

	i2 <- NewRoutingDecision(t1, a1, a2)
	h2 <- t1.Hash()
	WaitForIncomingDebt(t, p2, a1, owed)

	i1 <- NewRoutingDecision(t2, a1, a2)
	h1 <- t2.Hash()
	owed += t2.Amount()
	WaitForIncomingDebt(t, p1, a1, owed)
	WaitForOutgoingDebt(t, p1, a2, owed)
	i2 <- NewRoutingDecision(t2, a1, a2)
	h2 <- t2.Hash()
	WaitForIncomingDebt(t, p2, a1, owed)

	payment := testPayment{a1, a2, 4}
	p1.SendPayment(payment)
	owed -= 4
	WaitForIncomingDebt(t, p2, a1, owed)
	WaitForOutgoingDebt(t, p1, a2, owed)
	payment = testPayment{a1, a2, owed}
	p1.SendPayment(payment)
	WaitForIncomingDebt(t, p2, a1, 0)
	WaitForOutgoingDebt(t, p1, a2, 0)
}
