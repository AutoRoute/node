package node

import (
	"testing"
)

func TestSBRM(t *testing.T) {
	a := NodeAddress("1")
	s := NewBloomReachabilityMap()
	s.AddEntry(a)

	if !(s.IsReachable(a)) {
		t.Fatalf("expected %v to be reachable", s)
	}

	b := NewBloomReachabilityMap()

	err := b.Merge(s)
	if err != nil {
		t.Fatal(err)
	}

	if !(b.IsReachable(a)) {
		t.Fatalf("expected %v to be reachable", s)
	}
}
