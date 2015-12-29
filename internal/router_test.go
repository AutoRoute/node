package node

import (
	"log"
	"testing"
	"time"

	"github.com/AutoRoute/node/types"
)

func TestRouterConnections(t *testing.T) {
	sk1, _ := NewECDSAKey()
	k1 := sk1.PublicKey()
	sk2, _ := NewECDSAKey()
	k2 := sk2.PublicKey()
	r1 := NewRouter(k1)
	r2 := NewRouter(k2)
	defer r1.Close()
	defer r2.Close()
	Link(r1, r2)
	if len(r1.Connections()) != 1 {
		t.Fatal("Expected one connection in r1")
	}
	if len(r2.Connections()) != 1 {
		t.Fatal("Expected one connection in r2")
	}
}

func TestDoubleConnect(t *testing.T) {
	sk1, _ := NewECDSAKey()
	k1 := sk1.PublicKey()
	sk2, _ := NewECDSAKey()
	k2 := sk2.PublicKey()
	r1 := NewRouter(k1)
	r2 := NewRouter(k2)
	defer r1.Close()
	defer r2.Close()
	c1, c2 := MakePairedConnections(r1.GetAddress(), r2.GetAddress())
	r1.AddConnection(c2)
	r1.AddConnection(c2)
	r2.AddConnection(c1)
	r2.AddConnection(c1)

	if len(r1.Connections()) != 1 {
		t.Fatal("Expected one connection in r1")
	}
	if len(r2.Connections()) != 1 {
		t.Fatal("Expected one connection in r2")
	}
}

func TestDirectRouter(t *testing.T) {
	sk1, _ := NewECDSAKey()
	k1 := sk1.PublicKey()
	sk2, _ := NewECDSAKey()
	k2 := sk2.PublicKey()
	r1 := NewRouter(k1)
	r2 := NewRouter(k2)
	defer r1.Close()
	defer r2.Close()
	a2 := k2.Hash()
	Link(r1, r2)

	// Send a test packet over the connection
	p2 := testPacket(a2)
	go func() {
		err := r1.SendPacket(p2)
		if err != nil {
			t.Fatal(err)
		}
	}()

	received := <-r2.Packets()
	if received != p2 {
		t.Fatalf("%q != %q", received, p2)
	}

	// Make sure a bad packet fails
	a3 := types.NodeAddress("3")
	p3 := testPacket(a3)
	err := r1.SendPacket(p3)
	if err == nil {
		t.Fatal("Expected error got nil")
	}
}

func TestRelayRouter(t *testing.T) {
	sk1, err := NewECDSAKey()
	k1 := sk1.PublicKey()
	if err != nil {
		t.Fatal(err)
	}
	sk2, err := NewECDSAKey()
	k2 := sk2.PublicKey()
	if err != nil {
		t.Fatal(err)
	}
	sk3, err := NewECDSAKey()
	k3 := sk3.PublicKey()
	if err != nil {
		t.Fatal(err)
	}
	a3 := k3.Hash()
	r1 := NewRouter(k1)
	r2 := NewRouter(k2)
	r3 := NewRouter(k3)
	defer r1.Close()
	defer r2.Close()
	defer r3.Close()
	Link(r1, r2)
	Link(r2, r3)

	// Send a test packet over the connection
	p3 := testPacket(a3)
	go func() {
		tries := time.Tick(10 * time.Millisecond)
		timeout := time.After(time.Second)
		for {
			select {
			case <-tries:
				err := r1.SendPacket(p3)
				if err == nil {
					break
				}
			case <-timeout:
				log.Fatal("Timed out waiting for succesful send")
			}
		}
	}()

	received := <-r3.Packets()
	if received != p3 {
		t.Fatalf("%q != %q", received, p3)
	}
}
