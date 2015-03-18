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

	b := NewSimpleReachabilityMap()

	err := b.Merge(s)
	if err != nil {
		t.Fatal(err)
	}

	if !(b.IsReachable(a)) {
		t.Fatalf("expected %v to be reachable", s)
	}
}
