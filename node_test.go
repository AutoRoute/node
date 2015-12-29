package node

import (
	"testing"
	"time"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

func LinkRouters(a, b internal.Router) {
	c1, c2 := internal.MakePairedConnections(a.GetAddress(), b.GetAddress())
	a.AddConnection(c2)
	b.AddConnection(c1)
}

func TestfullNode(t *testing.T) {
	sk1, _ := internal.NewECDSAKey()
	sk2, _ := internal.NewECDSAKey()
	c := make(chan time.Time)
	n1 := newFullNode(sk1, internal.FakeMoney{}, time.Tick(100*time.Millisecond), c)
	n2 := newFullNode(sk2, internal.FakeMoney{}, time.Tick(100*time.Millisecond), time.Tick(100*time.Millisecond))
	defer n1.Close()
	defer n2.Close()
	internal.Link(n1, n2)

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
