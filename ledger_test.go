package node

import (
	"testing"
	"time"
)

func WaitForIncomingDebt(t *testing.T, l *ledger, a NodeAddress, o int64) {
	timeout := time.After(time.Second)
	tick := time.Tick(time.Millisecond)
	d := int64(0)
	for {
		select {
		case <-timeout:
			t.Errorf("Expected debt from %s to be %d != %d", a, o, d)
			return
		case <-tick:
			d, _ = l.IncomingDebt(a)
			if d == o {
				return
			}
		}
	}
}

func WaitForOutgoingDebt(t *testing.T, l *ledger, a NodeAddress, o int64) {
	timeout := time.After(time.Second)
	tick := time.Tick(time.Millisecond)
	d := int64(0)
	for {
		select {
		case <-timeout:
			t.Errorf("Expected debt from %s to be %d != %d", a, o, d)
			return
		case <-tick:
			d, _ = l.OutgoingDebt(a)
			if d == o {
				return
			}
		}
	}
}

func TestLedger(t *testing.T) {
	delivered := make(chan PacketHash)
	a1, a2, a3 := NodeAddress("1"), NodeAddress("2"), NodeAddress("3")

	routed := make(chan routingDecision)

	ledger := newLedger(a1, delivered, routed)

	t1 := testPacket(a2)
	t2 := testPacket(a3)
	routed <- newRoutingDecision(t1, a1, a2)
	routed <- newRoutingDecision(t2, a1, a2)

	owed := int64(0)
	WaitForIncomingDebt(t, ledger, a1, owed)
	WaitForOutgoingDebt(t, ledger, a2, owed)

	delivered <- t1.Hash()
	owed = t1.Amount()

	WaitForIncomingDebt(t, ledger, a1, owed)
	WaitForOutgoingDebt(t, ledger, a2, owed)

	delivered <- t2.Hash()
	owed += t2.Amount()

	WaitForIncomingDebt(t, ledger, a1, owed)
	WaitForOutgoingDebt(t, ledger, a2, owed)

	payment := Payment{a1, a2, 4}

	ledger.RecordPayment(payment)
	owed -= 4

	WaitForIncomingDebt(t, ledger, a1, owed)
	WaitForOutgoingDebt(t, ledger, a2, owed)

	payment = Payment{a1, a2, owed}
	ledger.RecordPayment(payment)
	owed -= owed

	WaitForIncomingDebt(t, ledger, a1, owed)
	WaitForOutgoingDebt(t, ledger, a2, owed)
}
