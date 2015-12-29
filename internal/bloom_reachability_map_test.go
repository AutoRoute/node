package node

import (
	"encoding/json"
	"testing"

	"github.com/AutoRoute/node/types"
)

func TestSBRM(t *testing.T) {
	a := types.NodeAddress("1")

	bloomMap1 := NewBloomReachabilityMap()
	bloomMap1.AddEntry(a)

	if !(bloomMap1.IsReachable(a)) {
		t.Fatalf("expected %s to be reachable in %v", a, bloomMap1)
	}

	// Test Merge
	bloomMap2 := NewBloomReachabilityMap()
	bloomMap2.Merge(bloomMap1)

	if !(bloomMap2.IsReachable(a)) {
		t.Fatalf("expected %s to be reachable in %v", a, bloomMap2)
	}

	// Test Merge after Increment
	bloomMap3 := NewBloomReachabilityMap()
	bloomMap3.Increment()
	bloomMap3.Merge(bloomMap2)

	if !(bloomMap3.IsReachable(a)) {
		t.Fatalf("expected %s to be reachable in %v", a, bloomMap3)
	}

	// Test Merge after Increment, other direction
	bloomMap2.Increment()
	bloomMap4 := NewBloomReachabilityMap()
	bloomMap4.Merge(bloomMap2)

	if !(bloomMap4.IsReachable(a)) {
		t.Fatalf("expected %s to be reachable in %v", a, bloomMap3)
	}
}

func TestBloomMarshalling(t *testing.T) {
	m := NewBloomReachabilityMap()
	m.AddEntry(types.NodeAddress("1"))
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var m2 BloomReachabilityMap
	err = json.Unmarshal(b, &m2)
	if err != nil {
		t.Fatal(err)
	}
}
