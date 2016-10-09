package integration_tests

import (
	"github.com/AutoRoute/node/types"

	"bytes"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"
)

// This tests that packets sent over a packet connection fully transit the system.
func TestPacket(t *testing.T) {
	listen := NewNodeBinary(BinaryOptions{
		Listen:       "[::1]:9999",
		FakeMoney:    true,
		Unix:         "/tmp/unix",
		RouteLogPath: "/tmp/route1.log",
	})
	listen.Start()
	defer listen.KillAndPrint(t)
	listen_id, err := WaitForID(listen)
	if err != nil {
		t.Fatal(err)
	}

	connect := NewNodeBinary(BinaryOptions{
		Listen:       "[::1]:9998",
		Connect:      []string{"[::1]:9999"},
		FakeMoney:    true,
		Unix:         "/tmp/unix2",
		RouteLogPath: "/tmp/route2.log",
	})
	connect.Start()
	defer connect.KillAndPrint(t)
	connect_id, err := WaitForID(connect)
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForConnection(listen, connect_id)
	if err != nil {
		t.Fatal(err)
	}

	raw_id, err := hex.DecodeString(connect_id)
	p := types.Packet{types.NodeAddress(string(raw_id)), 10, []byte("data")}

	c, err := WaitForSocket("/tmp/unix")
	if err != nil {
		t.Fatal(err)
	}
	c2, err := WaitForSocket("/tmp/unix2")
	if err != nil {
		t.Fatal(err)
	}

	w := json.NewEncoder(c)
	err = w.Encode(p)
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForPacketsReceived(listen, listen_id, 1)
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForPacketsSent(listen, connect_id, 1)
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForPacketsReceived(connect, listen_id, 1)
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForPacketsSent(connect, connect_id, 1)
	if err != nil {
		t.Fatal(err)
	}

	packets := make(chan types.Packet)
	go WaitForPacket(c2, t, packets)
	select {
	case <-time.After(4 * time.Second):
		t.Fatal("Never received packet")
	case p2 := <-packets:
		if p.Dest != p2.Dest || p.Amt != p2.Amt || !bytes.Equal(p.Data, p2.Data) {
			t.Fatal("Packets %v != %v", p, p2)
		}
	}
}
