package node

import (
	"testing"
	"time"
)

type testMapConnection struct {
	in  chan ReachabilityMap
	out chan ReachabilityMap
}

func (c testMapConnection) SendMap(m ReachabilityMap) error {
	c.out <- m
	return nil
}
func (c testMapConnection) ReachabilityMaps() <-chan ReachabilityMap {
	return c.in
}

func makePairedMapConnections() (MapConnection, MapConnection) {
	one := make(chan ReachabilityMap)
	two := make(chan ReachabilityMap)
	return testMapConnection{one, two}, testMapConnection{two, one}
}

func TestMapHandler(t *testing.T) {

	c1, c2 := makePairedMapConnections()

	a1 := NodeAddress("1")
	a2 := NodeAddress("2")

	m1 := newMapImpl(a1)
	m2 := newMapImpl(a1)

	m1.AddConnection(a2, c2)
	m2.AddConnection(a1, c1)

	timeout := func(m MapHandler, id NodeAddress) bool {
		start := time.Now()
		for now := start; now.Before(start.Add(time.Second)); now = time.Now() {
			if a, err := m.FindConnection(id); err == nil {
				if a == id {
					return true
				} else {
					t.Fatal("impossible?", a, id)
				}
			}
		}
		return false
	}

	if !timeout(m1, a2) {
		t.Fatal("timed out waiting for m1", m1)
	}
	if !timeout(m2, a1) {
		t.Fatal("timed out waiting for m1", m2)
	}
}

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

func TestRouter(t *testing.T) {

	a1 := NodeAddress("1")
	a2 := NodeAddress("2")

	k1 := pktest(a1)
	k2 := pktest(a2)

	r1 := newRouterImpl(a1)
	r2 := newRouterImpl(a2)

	c1, c2 := makePairedConnections(k1, k2)

	r1.AddConnection(c2)
	r2.AddConnection(c1)

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
