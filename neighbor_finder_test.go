package node

import (
	"fmt"
	"github.com/AutoRoute/l2"
	"testing"
)

type testInterface struct {
	in  chan l2.EthFrame
	out chan l2.EthFrame
}

func (t testInterface) ReadFrame() l2.EthFrame {
	return <-in
}

func (t testInterface) WriteFrame() (e l2.EthFrame) {
	out <- e
}

func CreatePairedInterface() (FrameReadWriter, FrameReadWriter) {
	one := make(chan l2.EthFrame)
	two := make(chan l2.EthFrame)
	return testInterface{one, two}, testInterface{two, one}
}

func TestBasicExchange(t *testing.T) {
	var test_mac1 string = "aa:bb:cc:dd:ee:00"
	var test_mac2 string = "aa:bb:cc:dd:ee:11"

	var public_key1 PublicKey = pktest("test1")
	var public_key2 PublicKey = pktest("test2")

	one, two := CreatePairedInterface()

	nf1 := NewNeighborFinder(public_key1)
	nf2 := NewNeighborFinder(public_key2)

	outone = nf1.Find(test_mac1, one)
	outtwo = nf2.Find(test_mac2, two)
}
