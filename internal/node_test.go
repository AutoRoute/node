package internal

import (
	"testing"
	"time"

	"github.com/AutoRoute/node/types"
)

func TestNode(t *testing.T) {
	sk1, _ := NewECDSAKey()
	sk2, _ := NewECDSAKey()
	c := make(chan time.Time)
	n1 := NewNode(sk1, FakeMoney{}, time.Tick(100*time.Millisecond), c)
	n2 := NewNode(sk2, FakeMoney{}, time.Tick(100*time.Millisecond), time.Tick(100*time.Millisecond))
	defer n1.Close()
	defer n2.Close()
	Link(n1, n2)

	if !n1.IsReachable(sk2.PublicKey().Hash()) {
		t.Fatalf("n2 is not reachable")
	}

	p2 := types.Packet{sk2.PublicKey().Hash(), 3, "data"}
	err := n1.SendPacket(p2)
	if err != nil {
		t.Fatalf("Error sending packet: %v", err)
	}
	<-n2.Packets()
}
