package node

import (
	"testing"
)

func TestSRM(t *testing.T) {
	a := NodeAddress("1")
	s := NewSimpleReachabilityMap()
	s.AddEntry(a)

	if !(s.IsReachable(a)) {
		t.Fatalf("expected %v to be reachable", s)
	}
}
