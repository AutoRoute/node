package node

import (
	"testing"
	"time"
)

func LinkRouters(a, b Router) {
	c1, c2 := makePairedConnections(a.GetAddress(), b.GetAddress())
	a.AddConnection(c2)
	b.AddConnection(c1)
}

func TestNode(t *testing.T) {
	sk1, _ := NewECDSAKey()
	sk2, _ := NewECDSAKey()
	c := make(chan time.Time)
	n1 := NewNode(sk1, time.Tick(100*time.Millisecond), c)
	n2 := NewNode(sk2, time.Tick(100*time.Millisecond), time.Tick(100*time.Millisecond))
	link(n1, n2)

	p2 := testPacket(sk2.PublicKey().Hash())
	n1.SendPacket(p2)
	<-n2.Packets()

	for range time.Tick(25 * time.Millisecond) {
		if amt, _ := n1.Router.OutgoingDebt(sk2.PublicKey().Hash()); amt != 0 {
			break
		}
	}

	c <- time.Now()

	for range time.Tick(25 * time.Millisecond) {
		if amt, _ := n1.Router.OutgoingDebt(sk2.PublicKey().Hash()); amt == 0 {
			return
		}
	}
}
