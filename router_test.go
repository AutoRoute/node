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

	go m1.AddConnection(a2, c2)
	go m2.AddConnection(a1, c1)

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
