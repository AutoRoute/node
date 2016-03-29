package integration_tests

import (
	"github.com/AutoRoute/node/types"

	"encoding/hex"
	"encoding/json"
	"testing"
	"time"
)

// This tests that packets sent over a packet connection fully transit the system.
func TestPacket(t *testing.T) {
	travis, err := CheckTravis()
	if err != nil {
		t.Logf("Warning: Checking for Travis failed: %s\n", err)
	}
	if travis {
		// TODO(danielp): Re-enable this once we figure out why it's hanging.
		t.Skip("TestPacket is temporarily disabled on Travis.")
	}

	listen := NewNodeBinary(BinaryOptions{
		Listen:     "[::1]:9999",
		Fake_money: true,
		Unix:       "/tmp/unix",
	}, true)
	listen.Start()
	defer listen.KillAndPrint(t)
	listen_id, err := WaitForID(listen)
	if err != nil {
		t.Fatal(err)
	}

	connect := NewNodeBinary(BinaryOptions{
		Listen:     "[::1]:9998",
		Connect:    []string{"[::1]:9999"},
		Fake_money: true,
		Unix:       "/tmp/unix2",
	}, true)
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
	p := types.Packet{types.NodeAddress(string(raw_id)), 10, "data"}

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
		if p != p2 {
			t.Fatal("Packets %v != %v", p, p2)
		}
	}
}
