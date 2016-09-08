package internal

import (
	"bytes"
	"testing"

	"github.com/AutoRoute/node/types"
)

func TestLogBloomFilter(t *testing.T) {
	buf := bytes.Buffer{}
	log := Logger{&buf}
	a := types.NodeAddress("1")

	bloomMap := NewBloomReachabilityMap()
	bloomMap.AddEntry(a)

	log.LogBloomFilter(bloomMap)

	bloomMap.Conglomerate.ReadFrom(&buf)

	if !(bloomMap.IsReachable(a)) {
		t.Fatal("expected %s to be reachable in %v", a, bloomMap)
	}
}
