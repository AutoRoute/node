package integration_tests

import (
	"github.com/AutoRoute/node"

	"encoding/hex"
	"encoding/json"
	"testing"
	"time"
)

// This tests that packets sent over a packet connection fully transit the system.
func TestPacket(t *testing.T) {
	listen := NewNodeBinary(BinaryOptions{
		Listen:     "[::1]:9999",
		Fake_money: true,
		Unix:       "/tmp/unix",
	})
	listen.Start()
	defer listen.KillAndPrint(t)
	_, err := WaitForID(listen)
	if err != nil {
		t.Fatal(err)
	}

	connect := NewNodeBinary(BinaryOptions{
		Listen:     "[::1]:9998",
		Connect:    []string{"[::1]:9999"},
		Fake_money: true,
		Unix:       "/tmp/unix2",
	})
	connect.Start()
	defer connect.KillAndPrint(t)
	connect_id, err := WaitForID(connect)

	err = WaitForConnection(listen, connect_id)
	if err != nil {
		t.Fatal(err)
	}

	raw_id, err := hex.DecodeString(connect_id)
	p := node.Packet{node.NodeAddress(string(raw_id)), 10, "data"}

	c, err := WaitForSocket("/tmp/unix")
	if err != nil {
		t.Fatal(err)
	}
	w := json.NewEncoder(c)
	err = w.Encode(p)
	if err != nil {
		t.Fatal(err)
	}

	c2, err := WaitForSocket("/tmp/unix2")
	if err != nil {
		t.Fatal(err)
	}
	packets := make(chan node.Packet)
	go WaitForPacket(c2, t, packets)
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("Never received packet")
	case p2 := <-packets:
		if p != p2 {
			t.Fatal("Packets %v != %v", p, p2)
		}
	}
}
