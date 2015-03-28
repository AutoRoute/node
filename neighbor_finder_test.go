package node

import (
	"github.com/AutoRoute/l2"
	"log"
	"testing"
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
	pk1 := pktest("test1")
	pk2 := pktest("test2")
	nf1 := NewNeighborData(pk1)
	nf2 := NewNeighborData(pk2)

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
	if msg != pk2.Hash() {
		log.Printf("Expected %q!=%q", pk2.Hash(), msg)
	}
	msg = <-outtwo
	if msg != pk1.Hash() {
		log.Printf("Expected %q!=%q", pk1.Hash(), msg)
	}
	msg = <-outone
	if msg != pk2.Hash() {
		log.Printf("Expected %q!=%q", pk2.Hash(), msg)
	}
	msg = <-outtwo
	if msg != pk1.Hash() {
		log.Printf("Expected %q!=%q", pk1.Hash(), msg)
	}
}
