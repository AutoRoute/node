package node

import (
	"testing"
	"time"
)

type testMapConnection struct {
	in  chan BloomReachabilityMap
	out chan BloomReachabilityMap
}

func (c testMapConnection) SendMap(m BloomReachabilityMap) error {
	c.out <- m
	return nil
}
func (c testMapConnection) ReachabilityMaps() <-chan BloomReachabilityMap {
	return c.in
}

func makePairedMapConnections() (MapConnection, MapConnection) {
	one := make(chan BloomReachabilityMap)
	two := make(chan BloomReachabilityMap)
	return testMapConnection{one, two}, testMapConnection{two, one}
}

func TestMapHandler(t *testing.T) {
	c1, c2 := makePairedMapConnections()
	a1 := NodeAddress("1")
	a2 := NodeAddress("2")
	m1 := newReachability(a1)
	m2 := newReachability(a1)
	m1.AddConnection(a2, c2)
	m2.AddConnection(a1, c1)

	timeout := func(m *reachabilityHandler, id NodeAddress) bool {
		start := time.Now()
		for now := start; now.Before(start.Add(time.Second)); now = time.Now() {
			if a, err := m.FindNextHop(id); err == nil {
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

func TestRelayMapHandler(t *testing.T) {
	c1, c2 := makePairedMapConnections()
	c3, c4 := makePairedMapConnections()
	a1 := NodeAddress("1")
	a2 := NodeAddress("2")
	a3 := NodeAddress("3")
	m1 := newReachability(a1)
	m2 := newReachability(a2)
	m3 := newReachability(a3)
	m1.AddConnection(a2, c2)
	m2.AddConnection(a1, c1)
	m2.AddConnection(a3, c3)
	m3.AddConnection(a2, c4)

	timeout := func(m *reachabilityHandler, id NodeAddress, nexthop NodeAddress) bool {
		start := time.Now()
		for now := start; now.Before(start.Add(time.Second)); now = time.Now() {
			if a, err := m.FindNextHop(id); err == nil {
				if a == nexthop {
					return true
				} else {
					t.Fatal("Wrong nexthop?", a, id)
				}
			}
		}
		return false
	}

	if !timeout(m3, a2, a2) {
		t.Fatal("timed out waiting for m1", m3)
	}
	if !timeout(m3, a1, a2) {
		t.Fatal("timed out waiting for m1", m3)
	}
}
