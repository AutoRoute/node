package internal

import (
	"testing"
	"time"

	"github.com/AutoRoute/node/types"
)

func TestMapHandler(t *testing.T) {
	c1, c2 := makePairedMapConnections()
	a1 := types.NodeAddress("1")
	a2 := types.NodeAddress("2")
	lgr1 := &testLogger{0, 0}
	lgr2 := &testLogger{0, 0}
	m1 := newReachability(a1, lgr1)
	m2 := newReachability(a1, lgr2)
	defer m1.Close()
	defer m2.Close()
	m1.AddConnection(a2, c2)
	m2.AddConnection(a1, c1)

	timeout := func(m *reachabilityHandler, id types.NodeAddress) bool {
		start := time.Now()
		for now := start; now.Before(start.Add(time.Second)); now = time.Now() {
			if nodes, err := m.FindPossibleDests(id, ""); err == nil {
				for _, node := range nodes {
					if node == id {
						return true
					}
				}
				t.Fatal("impossible?", nodes, id)
			}
		}
		return false
	}

	if !timeout(m1, a2) {
		t.Fatal("Timed out waiting for m1", m1)
	}
	if !timeout(m2, a1) {
		t.Fatal("Timed out waiting for m1", m2)
	}
	if lgr1.BloomCount != 1 || lgr2.BloomCount != 1 {
		t.Fatal("Not all connections logged", lgr1.BloomCount, lgr2.BloomCount)
	}
}

func TestRelayMapHandler(t *testing.T) {
	c1, c2 := makePairedMapConnections()
	c3, c4 := makePairedMapConnections()
	a1 := types.NodeAddress("1")
	a2 := types.NodeAddress("2")
	a3 := types.NodeAddress("3")
	lgr1 := &testLogger{0, 0}
	lgr2 := &testLogger{0, 0}
	lgr3 := &testLogger{0, 0}
	m1 := newReachability(a1, lgr1)
	m2 := newReachability(a2, lgr2)
	m3 := newReachability(a3, lgr3)
	defer m1.Close()
	defer m2.Close()
	defer m3.Close()
	m1.AddConnection(a2, c2)
	m2.AddConnection(a1, c1)
	m2.AddConnection(a3, c3)
	m3.AddConnection(a2, c4)

	timeout := func(m *reachabilityHandler, id types.NodeAddress, nexthop types.NodeAddress) bool {
		start := time.Now()
		for now := start; now.Before(start.Add(time.Second)); now = time.Now() {
			if nodes, err := m.FindPossibleDests(id, ""); err == nil {
				for _, node := range nodes {
					if node == nexthop {
						return true
					}
				}
				t.Fatal("Wrong nexthop?", nodes, id)
			}
		}
		return false
	}

	if !timeout(m3, a2, a2) {
		t.Fatal("Timed out waiting for m1", m3)
	}
	if !timeout(m3, a1, a2) {
		t.Fatal("Timed out waiting for m1", m3)
	}
	if lgr1.BloomCount != 1 || lgr2.BloomCount != 2 || lgr3.BloomCount != 1 {
		t.Fatal("Not all connections logged", lgr1.BloomCount, lgr2.BloomCount, lgr3.BloomCount)
	}
}

// Make sure it can't send packets backwards.
func TestNoBackwardsPackets(t *testing.T) {
	conn1, conn2 := makePairedMapConnections()
	// We don't actually care about whether conn3 gets maps or not, because we're
	// just using it as a destination.
	conn3, _ := makePairedMapConnections()

	address1 := types.NodeAddress("1")
	address2 := types.NodeAddress("2")
	address3 := types.NodeAddress("3")

	lgr1 := &testLogger{0, 0}
	lgr2 := &testLogger{0, 0}

	reach1 := newReachability(address1, lgr1)
	reach2 := newReachability(address2, lgr2)
	defer reach1.Close()
	defer reach2.Close()

	reach1.AddConnection(address2, conn2)
	reach2.AddConnection(address1, conn1)
	reach2.AddConnection(address3, conn3)

	// If we tell it the packet came from conn2, it should refuse to send it along
	// to conn2, even though it could.
	dest, err := reach1.FindPossibleDests(address3, address2)
	if dest != nil {
		t.Fatalf("Expected empty dest, got '%s'.\n", dest)
	}
	if err == nil {
		t.Fatal("Expected error, but got none.")
	}
	if lgr1.BloomCount != 1 || lgr2.BloomCount != 2 {
		t.Fatal("Not all connectinos logged", lgr1.BloomCount, lgr2.BloomCount)
	}
}

// Test that multiply sending the same map doesn't generate multiple relays
func TestNoMapCycles(t *testing.T) {
	conn1, conn3 := makePairedMapConnections()
	conn2, conn4 := makePairedMapConnections()

	address1 := types.NodeAddress("1")
	address2 := types.NodeAddress("2")
	address3 := types.NodeAddress("3")
	addressSent := types.NodeAddress("Sent")
	addressSent2 := types.NodeAddress("Sent2")

	lgr := &testLogger{0, 0}

	reach1 := newReachability(address1, lgr)
	defer reach1.Close()

	reach1.AddConnection(address2, conn3)
	<-conn1.ReachabilityMaps()
	reach1.AddConnection(address3, conn4)
	<-conn2.ReachabilityMaps()

	m := NewBloomReachabilityMap()
	m.AddEntry(addressSent)
	conn1.SendMap(m.Copy())

	found_address := false
	timeout := time.After(10 * time.Second)
loop:
	for {
		select {
		case c := <-conn2.ReachabilityMaps():
			if c.IsReachable(addressSent) {
				found_address = true
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	if !found_address {
		t.Fatal("Cannot find find sent address in received map")
	}

	found_address = false
	found_address2 := false
	go conn1.SendMap(m.Copy())
	m2 := NewBloomReachabilityMap()
	m2.AddEntry(addressSent2)
	go conn1.SendMap(m2.Copy())
	timeout = time.After(10 * time.Second)

loop2:
	for {
		select {
		case c := <-conn2.ReachabilityMaps():
			if c.IsReachable(addressSent) {
				found_address = true
			}
			if c.IsReachable(addressSent2) {
				found_address2 = true
				break loop2
			}
		case <-timeout:
			break loop2
		}
	}
	if found_address {
		t.Fatal("Map was sent when it should not have been")
	}
	if !found_address2 {
		t.Fatal("Second address was not sent")
	}
	if lgr.BloomCount != 2 {
		t.Fatal("Not all connections logged", lgr.BloomCount)
	}
}
