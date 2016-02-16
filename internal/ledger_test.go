package internal

import (
	"testing"
	"time"

	"github.com/AutoRoute/node/types"
)

func WaitForIncomingDebt(t *testing.T, l *Ledger, a types.NodeAddress, o int64) {
	timeout := time.After(time.Second)
	tick := time.Tick(time.Millisecond)
	d := int64(0)
	for {
		select {
		case <-timeout:
			t.Errorf("Expected incoming debt from %s to be %d != %d", a, o, d)
			return
		case <-tick:
			d = l.IncomingDebt(a)
			if d == o {
				return
			}
		}
	}
}

func WaitForOutgoingDebt(t *testing.T, l *Ledger, a types.NodeAddress, o int64) {
	timeout := time.After(time.Second)
	tick := time.Tick(time.Millisecond)
	d := int64(0)
	for {
		select {
		case <-timeout:
			t.Errorf("Expected outgoing debt from %s to be %d != %d", a, o, d)
			return
		case <-tick:
			d = l.OutgoingDebt(a)
			if d == o {
				return
			}
		}
	}
}

func TestLedger(t *testing.T) {
	delivered := make(chan types.PacketHash)
	a1, a2, a3 := types.NodeAddress("1"), types.NodeAddress("2"), types.NodeAddress("3")

	routed := make(chan routingDecision)

	Ledger := newLedger(a1, delivered, routed)
	defer Ledger.Close()

	// two packets from a1 to a2
	t1 := testPacket(a2)
	t2 := testPacket(a3)
	routed <- newRoutingDecision(t1, a1, a2)
	routed <- newRoutingDecision(t2, a1, a2)

	owed := int64(0)
	WaitForIncomingDebt(t, Ledger, a1, owed)
	WaitForOutgoingDebt(t, Ledger, a2, owed)

	// t1 is delivered
	delivered <- t1.Hash()
	owed = t1.Amount()

	WaitForIncomingDebt(t, Ledger, a1, owed)
	WaitForOutgoingDebt(t, Ledger, a2, owed)

	// t2 is delievered.
	delivered <- t2.Hash()
	owed += t2.Amount()

	WaitForIncomingDebt(t, Ledger, a1, owed)
	WaitForOutgoingDebt(t, Ledger, a2, owed)

	// We pay for some of it.
	c := make(chan bool, 1)
	c <- true
	Ledger.RecordPayment(a2, 4, c)
	owed -= 4

	WaitForIncomingDebt(t, Ledger, a1, owed)
	WaitForOutgoingDebt(t, Ledger, a2, owed)

	// We send a payment which is never accepted.
	c <- false
	Ledger.RecordPayment(a2, 4, c)
	WaitForIncomingDebt(t, Ledger, a1, owed)
	WaitForOutgoingDebt(t, Ledger, a2, owed)

	// We pay for the rest.
	c <- true
	Ledger.RecordPayment(a2, owed, c)
	owed -= owed

	WaitForIncomingDebt(t, Ledger, a1, owed)
	WaitForOutgoingDebt(t, Ledger, a2, owed)
}
