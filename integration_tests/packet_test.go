package integration_tests

import (
	"github.com/AutoRoute/node"

	"encoding/hex"
	"encoding/json"
	"net"
	"testing"
	"time"
)

func WaitForSocket(p string) (net.Conn, error) {
	timeout := time.After(time.Second)
	for range time.Tick(10 * time.Millisecond) {
		c, err := net.Dial("unix", "/tmp/unix")
		if err == nil {
			return c, err
		}
		select {
		case <-timeout:
			return c, err
		default:
		}
	}
	panic("Unreachable")
}

func WaitForPacket(c net.Conn, t *testing.T, s chan node.Packet) {
	r := json.NewDecoder(c)
	var p node.Packet
	err := r.Decode(&p)
	if err != nil {
		t.Fatal(err)
	}
	s <- p
}

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
