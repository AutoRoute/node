package internal

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/AutoRoute/node/types"
)

func TestLogBloomFilter(t *testing.T) {
	buf := bytes.Buffer{}
	logger := NewLogger(&buf)
	a := types.NodeAddress("1")

	bloomMap1 := NewBloomReachabilityMap()
	bloomMap1.AddEntry(a)

	logger.LogBloomFilter(bloomMap1)

	var bloomMap2 BloomReachabilityMap
	decoder := json.NewDecoder(&buf)
	err := decoder.Decode(&bloomMap2.Conglomerate)
	if err != nil {
		t.Fatal(err)
	}

	if !(bloomMap2.IsReachable(a)) {
		t.Fatal("expected %s to be reachable in %v", a, bloomMap2)
	}
}
