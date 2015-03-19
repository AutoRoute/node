package node

import (
	"fmt"
	"github.com/AutoRoute/l2"
	"sync"
	"testing"
)

func TestNetExchange(t *testing.T) {
	fmt.Println("Testing Basic Net Exchange")
	var test_mac1 string = "aa:bb:cc:dd:ee:00"
	var test_mac2 string = "aa:bb:cc:dd:ee:11"

	var public_key1 PublicKey = pktest("test1")
	var public_key2 PublicKey = pktest("test2")

	one, errdev1 := l2.NewTapDevice(test_mac1, "one")
	if errdev1 != nil {
		panic(errdev1)
	}
	two, errdev2 := l2.NewTapDevice(test_mac2, "two")
	if errdev2 != nil {
		panic(errdev2)
	}

	nf1 := NewNeighborData(public_key1)
	nf2 := NewNeighborData(public_key2)

	outone, errfind1 := nf1.Find(test_mac1, one)
	if errfind1 != nil {
		panic(errfind1)
	}
	outtwo, errfind2 := nf2.Find(test_mac2, two)
	if errfind2 != nil {
		panic(errfind2)
	}
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(2)
		go CheckReceivedMessage(outone, string(public_key2.Hash()), string(public_key1.Hash()), &wg)
		go CheckReceivedMessage(outtwo, string(public_key1.Hash()), string(public_key2.Hash()), &wg)
		wg.Wait()
	}
	fmt.Println("Done")
}
