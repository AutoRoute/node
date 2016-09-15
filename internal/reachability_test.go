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
	m1 := newReachability(a1, testLogger{})
	m2 := newReachability(a1, testLogger{})
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
		t.Fatal("timed out waiting for m1", m1)
	}
	if !timeout(m2, a1) {
		t.Fatal("timed out waiting for m1", m2)
	}
}

func TestRelayMapHandler(t *testing.T) {
	c1, c2 := makePairedMapConnections()
	c3, c4 := makePairedMapConnections()
	a1 := types.NodeAddress("1")
	a2 := types.NodeAddress("2")
	a3 := types.NodeAddress("3")
	m1 := newReachability(a1, testLogger{})
	m2 := newReachability(a2, testLogger{})
	m3 := newReachability(a3, testLogger{})
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
		t.Fatal("timed out waiting for m1", m3)
	}
	if !timeout(m3, a1, a2) {
		t.Fatal("timed out waiting for m1", m3)
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

	reach1 := newReachability(address1, testLogger{})
	reach2 := newReachability(address2, testLogger{})
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
		t.Fatal("Expected error, but got none.\n")
	}
}
