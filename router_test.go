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
	k PublicKey
}

func (t testConnection) Close() error   { return nil }
func (t testConnection) Key() PublicKey { return t.k }

func makePairedConnections(k1, k2 PublicKey) (Connection, Connection) {
	d1, d2 := makePairedDataConnections()
	m1, m2 := makePairedMapConnections()

	return testConnection{d1, m1, k1}, testConnection{d2, m2, k2}
}

type testPacket NodeAddress

func (p testPacket) Destination() NodeAddress { return NodeAddress(p) }

func LinkRouters(a, b Router) {
	c1, c2 := makePairedConnections(a.GetAddress(), b.GetAddress())
	a.AddConnection(c2)
	b.AddConnection(c1)
}

func TestDirectRouter(t *testing.T) {

	a1 := NodeAddress("1")
	a2 := NodeAddress("2")
	r1 := newRouterImpl(a1)
	r2 := newRouterImpl(a2)
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
	err := r1.SendPacket(p3)
	if err == nil {
		t.Fatal("Expected error got nil")
	}
}

func TestRelayRouter(t *testing.T) {

	a1 := NodeAddress("1")
	a2 := NodeAddress("2")
	a3 := NodeAddress("3")
	r1 := newRouterImpl(a1)
	r2 := newRouterImpl(a2)
	r3 := newRouterImpl(a3)
	LinkRouters(r1, r2)
	LinkRouters(r2, r3)

	// Send a test packet over the connection
	p3 := testPacket(a3)
	go func() {
		start := time.Now()
		var err error = nil
		for now := start; now.Before(start.Add(time.Second)); now = time.Now() {
			err = r1.SendPacket(p3)
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Fatal(err)
		}
	}()

	received := <-r3.Packets()
	if received != p3 {
		t.Fatalf("%q != %q", received, p3)
	}
}
