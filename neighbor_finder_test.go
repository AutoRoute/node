package node

import (
	"github.com/AutoRoute/l2"
	"log"
	"sync"
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

func CheckReceivedMessage(cs <-chan string, test string, wg *sync.WaitGroup) {
	defer wg.Done()
	msg := <-cs
	log.Printf("received %q", msg)
	if msg != test {
		log.Fatalf("%Received %q != %q", msg, test)
	}
}

func TestBasicExchange(t *testing.T) {
	test_mac1, _ := l2.MacToBytes("aa:bb:cc:dd:ee:00")
	test_mac2, _ := l2.MacToBytes("aa:bb:cc:dd:ee:11")

	public_key1 := pktest("test1")
	public_key2 := pktest("test2")

	one, two := CreatePairedInterface()

	wg := &sync.WaitGroup{}
	wg.Add(4)

	nf1 := NewNeighborData(public_key1)
	nf2 := NewNeighborData(public_key2)

	outone, err := nf1.Find(test_mac1, one)
	if err != nil {
		t.Fatal(err)
	}
	outtwo, err := nf2.Find(test_mac2, two)
	if err != nil {
		t.Fatal(err)
	}
	// We should receive the other side twice, once from broadcast, and one
	// from directed response
	go CheckReceivedMessage(outone, string(public_key2.Hash()), wg)
	go CheckReceivedMessage(outone, string(public_key2.Hash()), wg)
	go CheckReceivedMessage(outtwo, string(public_key1.Hash()), wg)
	go CheckReceivedMessage(outtwo, string(public_key1.Hash()), wg)
	wg.Wait()
}
