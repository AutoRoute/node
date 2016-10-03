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
	logger := NewLogger(&buf)
	a := types.NodeAddress("1")

	bloomMap1 := internal.NewBloomReachabilityMap()
	bloomMap1.AddEntry(a)

	err := logger.LogBloomFilter(bloomMap1)
	if err != nil {
		t.Fatal(err)
	}

	var bloomMap2 internal.BloomReachabilityMap
	decoder := json.NewDecoder(&buf)
	err = decoder.Decode(&bloomMap2.Conglomerate)
	if err != nil {
		t.Fatal(err)
	}

	if !(bloomMap2.IsReachable(a)) {
		t.Fatal("expected %s to be reachable in %v", a, bloomMap2)
	}
}

func TestLogRoutingDecision(t *testing.T) {
	buf := bytes.Buffer{}
	logger := NewLogger(&buf)
	dest := types.NodeAddress("destination")
	next := types.NodeAddress("next_hop")
	packet_size := 10
	amt := int64(7)

	err := logger.LogRoutingDecision(dest, next, packet_size, amt)
	if err != nil {
		t.Fatal(err)
	}

	var rd routingDecision
	decoder := json.NewDecoder(&buf)
	err = decoder.Decode(&rd)
	if err != nil {
		t.Fatal(err)
	}

	if rd.Dest != dest || rd.Next != next ||
		rd.PacketSize != packet_size || rd.Amt != amt {
		t.Fatal("Logger didn't couldn't decode log entry", rd.Dest, rd.Next, rd.PacketSize, rd.Amt)
	}
}
