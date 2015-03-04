package node

import (
	"testing"
)

func TestSRM(t *testing.T) {
	a := NodeAddress("1")
	m := make(map[NodeAddress]bool)
	m[a] = true
	s := SimpleReachabilityMap(m)

	if !(s.IsReachable(a)) {
		t.Fatalf("expected %q to be reachable", s)
	}
}
