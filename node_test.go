package node

import (
	"testing"
	"time"
)

func LinkRouters(a, b router) {
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
	defer n1.Close()
	defer n2.Close()
	link(n1, n2)

	if !n1.IsReachable(sk2.PublicKey().Hash()) {
		t.Fatalf("n2 is not reachable")
	}

	p2 := testPacket(sk2.PublicKey().Hash())
	err := n1.SendPacket(p2)
	if err != nil {
		t.Fatalf("Error sending packet: %v", err)
	}
	<-n2.Packets()

	for range time.Tick(25 * time.Millisecond) {
		if amt, _ := n1.router.OutgoingDebt(sk2.PublicKey().Hash()); amt != 0 {
			break
		}
	}

	c <- time.Now()

	for range time.Tick(25 * time.Millisecond) {
		if amt, _ := n1.router.OutgoingDebt(sk2.PublicKey().Hash()); amt == 0 {
			return
		}
	}
}
