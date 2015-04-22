package node

import (
	"testing"
)

func TestSBRM(t *testing.T) {
	a := NodeAddress("1")

	bloomMap1 := NewBloomReachabilityMap()
	bloomMap1.AddEntry(a)

	if !(bloomMap1.IsReachable(a)) {
		t.Fatalf("expected %s to be reachable in %v", a, bloomMap1)
	}

	// Test Merge
	bloomMap2 := NewBloomReachabilityMap()
	err1 := bloomMap2.Merge(bloomMap1)
	if err1 != nil {
		t.Fatal(err1)
	}

	if !(bloomMap2.IsReachable(a)) {
		t.Fatalf("expected %s to be reachable in %v", a, bloomMap2)
	}

	// Test Merge after Increment
	bloomMap3 := NewBloomReachabilityMap()
	bloomMap3.Increment()
	err2 := bloomMap3.Merge(bloomMap2)
	if err2 != nil {
		t.Fatal("A merge across different distance SBRM's failed")
	}

	if !(bloomMap3.IsReachable(a)) {
		t.Fatalf("expected %s to be reachable in %v", a, bloomMap3)
	}

}
