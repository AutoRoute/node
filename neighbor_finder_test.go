package node

import (
	"fmt"
	"github.com/AutoRoute/l2"
	"log"
	"testing"
	"time"
)

type testInterface struct {
	in  chan l2.EthFrame
	out chan l2.EthFrame
}

func (t testInterface) ReadFrame() (l2.EthFrame, error) {
	return <-t.in, nil // TODO: return error
}

func (t testInterface) WriteFrame(e l2.EthFrame) error {
	t.out <- e
	return nil // TODO: return error
}

func CreatePairedInterface() (l2.FrameReadWriter, l2.FrameReadWriter) {
	one := make(chan l2.EthFrame, 100)
	two := make(chan l2.EthFrame, 100)
	return testInterface{one, two}, testInterface{two, one}
}

func CheckReceivedMessage(cs <-chan string, test string, receiver string) {
	var msg string = <-cs
	if msg != test {
		log.Fatalf("%q: Received %q != %q", receiver, msg, test)
	}
	fmt.Printf("%q: Received: %v\n", receiver, msg)
}

func TestBasicExchange(t *testing.T) {
	var test_mac1 string = "aa:bb:cc:dd:ee:00"
	var test_mac2 string = "aa:bb:cc:dd:ee:11"

	var public_key1 PublicKey = pktest("test1")
	var public_key2 PublicKey = pktest("test2")

	one, two := CreatePairedInterface()

	nf1 := NewNeighborData(public_key1)
	nf2 := NewNeighborData(public_key2)

	var outone <-chan string
	var outtwo <-chan string
	var err1 error
	var err2 error
	go func() {
		outone, err1 = nf1.Find(test_mac1, one)
		if err1 != nil {
			panic(err1)
		}
	}()
	go func() {
		outtwo, err2 = nf2.Find(test_mac2, two)
		if err2 != nil {
			panic(err2)
		}
	}()
	for i := 0; i < 100; i++ {
		go CheckReceivedMessage(outone, string(public_key2.Hash()), string(public_key1.Hash()))
		go CheckReceivedMessage(outtwo, string(public_key1.Hash()), string(public_key2.Hash()))
		time.Sleep(1 * 1e9)
	}
}
