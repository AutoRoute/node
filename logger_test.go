package node

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

func TestLogBloomFilter(t *testing.T) {
	buf := bytes.Buffer{}
	lgr := NewLogger(&buf)
	a := types.NodeAddress("1")

	bloomMap1 := internal.NewBloomReachabilityMap()
	bloomMap1.AddEntry(a)

	err := lgr.LogBloomFilter(bloomMap1)
	if err != nil {
		t.Fatal(err)
	}

	var bloomMap2 internal.BloomReachabilityMap
	dec := json.NewDecoder(&buf)
	err = dec.Decode(&bloomMap2.Conglomerate)
	if err != nil {
		t.Fatal(err)
	}

	if !(bloomMap2.IsReachable(a)) {
		t.Fatal("Expected %s to be reachable in %v", a, bloomMap2)
	}
}

func TestRoutingDecision(t *testing.T) {
	buf := bytes.Buffer{}
	lgr := NewLogger(&buf)
	dest := types.NodeAddress("destination")
	next := types.NodeAddress("next_hop")
	packet_size := 10
	amt := int64(7)

	err := lgr.LogRoutingDecision(dest, next, packet_size, amt)
	if err != nil {
		t.Fatal(err)
	}

	var rd routingDecision
	dec := json.NewDecoder(&buf)
	err = dec.Decode(&rd)
	if err != nil {
		t.Fatal(err)
	}

	if rd.Dest != dest || rd.Next != next ||
		rd.PacketSize != packet_size || rd.Amt != amt {
		t.Fatal("Unexpected log entry", rd.Dest, rd.Next, rd.PacketSize, rd.Amt)
	}
}
