package node

import (
	"testing"
)

func TestSBRM(t *testing.T) {
	a := NodeAddress("1")
	s := NewSimpleBloomReachabilityMap()
	s.AddEntry(a)

	if !(s.IsReachable(a)) {
		t.Fatalf("expected %v to be reachable", s)
	}

	b := NewSimpleBloomReachabilityMap()
	b.Merge(s)
	if !(b.IsReachable(a)) {
		t.Fatalf("expected %v to be reachable", s)
	}
}
