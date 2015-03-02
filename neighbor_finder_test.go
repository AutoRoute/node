package node

import (
	"fmt"
	"testing"
)

type testInterface struct {
	in chan EthFrame
	out chan EthFrame
}

func (t testInterface) ReadFrame() EthFrame {
	return <- in
}

func (t testInterface) WriteFrame() (e EthFrame) {
	out <- e
}

func CreatePairedInterface() (FrameReadWriter, FrameReadWriter) {
	one := make(chan EthFrame)
	two := make(chan EthFrame)
	return testInterface{one,two}, testInterface{two, one}
}

func NewNeighborFinder(pk PublicKey) NeighborFinder {
	return layertwo{pk}
}

func TestBasicExchange(t *testing.T) {
	one, two := CreatePairedInterface()

	nf1 := layertwo{public_key1}
	nf2 := NewNeighborFinder(public_key2)

	outone := nf1.Find(test_mac1, one)
	outtwo := nf2.Find(test_mac2, two)
}
