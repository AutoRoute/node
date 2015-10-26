package node

import (
	"net"
	"testing"

	"github.com/AutoRoute/l2"
)

type testInterface struct {
	in  chan l2.EthFrame
	out chan l2.EthFrame
}

func (t testInterface) ReadFrame() (l2.EthFrame, error) {
	return <-t.in, nil
}

func (t testInterface) WriteFrame(e l2.EthFrame) error {
	t.out <- e
	return nil
}

func CreatePairedInterface() (l2.FrameReadWriter, l2.FrameReadWriter) {
	// Make non blocking since buffering exists on ethernet drivers, so
	// we can be less stringent
	one := make(chan l2.EthFrame, 1)
	two := make(chan l2.EthFrame, 1)
	return testInterface{one, two}, testInterface{two, one}
}

func TestBasicExchange(t *testing.T) {

	test_mac1, _ := l2.MacToBytes("aa:bb:cc:dd:ee:00")
	test_mac2, _ := l2.MacToBytes("aa:bb:cc:dd:ee:11")

	k1, err := NewECDSAKey()
	pk1 := k1.PublicKey()
	if err != nil {
		t.Fatal(err)
	}
	k2, err := NewECDSAKey()
	pk2 := k2.PublicKey()
	if err != nil {
		t.Fatal(err)
	}

	ll_addr_str1 := "fe80::11"
	ll_addr_str2 := "fe80::22"
	ll_addr1 := net.ParseIP(ll_addr_str1)
	if ll_addr1 == nil {
		t.Fatal("Unable to parse IP address")
	}
	ll_addr2 := net.ParseIP(ll_addr_str2)
	if ll_addr2 == nil {
		t.Fatal("Unable to parse IP address")
	}
	port1 := uint16(34321)
	port2 := uint16(34322)

	nf1 := NewNeighborFinder(pk1, ll_addr1, port1)
	nf2 := NewNeighborFinder(pk2, ll_addr2, port2)

	one, two := CreatePairedInterface()
	outone, err := nf1.Find(test_mac1, one)
	if err != nil {
		t.Fatal(err)
	}
	outtwo, err := nf2.Find(test_mac2, two)
	if err != nil {
		t.Fatal(err)
	}

	// We should receive the other side twice, once from broadcast, and once
	// from directed response
	msg := <-outone
	if msg.NodeAddr != pk2.Hash() {
		t.Errorf("Expected %q!=%q", pk2.Hash(), msg.NodeAddr)
	}
	if msg.LLAddrStr != ll_addr_str2 {
		t.Errorf("Expected %q!=%q", ll_addr_str2, msg.LLAddrStr)
	}
  if msg.Port != port2 {
    t.Errorf("Expected %q!=%q", port2, msg.Port)
  }

	msg = <-outtwo
	if msg.NodeAddr != pk1.Hash() {
		t.Errorf("Expected %q!=%q", pk1.Hash(), msg.NodeAddr)
	}
	if msg.LLAddrStr != ll_addr_str1 {
		t.Errorf("Expected %q!=%q", ll_addr_str1, msg.LLAddrStr)
	}
  if msg.Port != port1 {
    t.Errorf("Expected %q!=%q", port1, msg.Port)
  }

	msg = <-outone
	if msg.NodeAddr != pk2.Hash() {
		t.Errorf("Expected %q!=%q", pk2.Hash(), msg.NodeAddr)
	}
	if msg.LLAddrStr != ll_addr_str2 {
		t.Errorf("Expected %q!=%q", ll_addr_str2, msg.LLAddrStr)
	}
  if msg.Port != port2 {
    t.Errorf("Expected %q!=%q", port2, msg.Port)
  }

	msg = <-outtwo
	if msg.NodeAddr != pk1.Hash() {
		t.Errorf("Expected %q!=%q", pk1.Hash(), msg.NodeAddr)
	}
	if msg.LLAddrStr != ll_addr_str1 {
		t.Errorf("Expected %q!=%q", ll_addr_str1, msg.LLAddrStr)
	}
  if msg.Port != port1 {
    t.Errorf("Expected %q!=%q", port1, msg.Port)
  }
}
