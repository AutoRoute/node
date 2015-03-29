package node

import (
	"log"
	"testing"
	"time"
)

type testDataConnection struct {
	in  chan Packet
	out chan Packet
}

func (c testDataConnection) SendPacket(m Packet) error {
	c.out <- m
	return nil
}
func (c testDataConnection) Packets() <-chan Packet {
	return c.in
}

func makePairedDataConnections() (DataConnection, DataConnection) {
	one := make(chan Packet)
	two := make(chan Packet)
	return testDataConnection{one, two}, testDataConnection{two, one}
}

type testConnection struct {
	DataConnection
	MapConnection
	ReceiptConnection
	k PublicKey
}

func (t testConnection) Close() error   { return nil }
func (t testConnection) Key() PublicKey { return t.k }

func makePairedConnections(k1, k2 PublicKey) (Connection, Connection) {
	d1, d2 := makePairedDataConnections()
	m1, m2 := makePairedMapConnections()
	r1, r2 := makePairedReceiptConnections()
	return testConnection{d1, m1, r1, k1}, testConnection{d2, m2, r2, k2}
}

type testPacket NodeAddress

func (p testPacket) Destination() NodeAddress { return NodeAddress(p) }
func (p testPacket) Hash() PacketHash         { return PacketHash(p) }
func (p testPacket) Amount() int64            { return 1 }

func LinkRouters(a, b Router) {
	c1, c2 := makePairedConnections(a.GetAddress(), b.GetAddress())
	a.AddConnection(c2)
	b.AddConnection(c1)
}

func TestDirectRouter(t *testing.T) {
	k1, err := NewPublicKey()
	if err != nil {
		t.Fatal(err)
	}
	k2, err := NewPublicKey()
	if err != nil {
		t.Fatal(err)
	}
	r1 := newRouterImpl(k1)
	r2 := newRouterImpl(k2)
	a2 := k2.Hash()
	LinkRouters(r1, r2)

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
	a3 := NodeAddress("3")
	p3 := testPacket(a3)
	err = r1.SendPacket(p3)
	if err == nil {
		t.Fatal("Expected error got nil")
	}
}

func TestRelayRouter(t *testing.T) {
	k1, err := NewPublicKey()
	if err != nil {
		t.Fatal(err)
	}
	k2, err := NewPublicKey()
	if err != nil {
		t.Fatal(err)
	}
	k3, err := NewPublicKey()
	if err != nil {
		t.Fatal(err)
	}
	a3 := k3.Hash()
	r1 := newRouterImpl(k1)
	r2 := newRouterImpl(k2)
	r3 := newRouterImpl(k3)
	LinkRouters(r1, r2)
	LinkRouters(r2, r3)

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
