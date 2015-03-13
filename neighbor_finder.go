package node

import (
	"bytes"
	"fmt"
	"github.com/AutoRoute/l2"
	"log"
)

// The layer two protocol takes a layer two device and returns the hash of the
// Public Key of all neighbors it can find.
type NeighborFinder interface {
	Find(l2.FrameReadWriter) <-chan string
}

type NeighborData struct {
	pk PublicKey
}

func NewNeighborData(pk PublicKey) NeighborData {
	return NeighborData{pk}
}

func (n NeighborData) Find(mac string, frw l2.FrameReadWriter) (<-chan string, error) {
	c := make(chan string)
	// Broadcast Hash
	broadcastAddr, errb := l2.MacToBytes("ff:ff:ff:ff:ff:ff")
	if errb != nil {
		log.Fatalf("%v\n", errb)
	}
	localAddr, errl := l2.MacToBytes(mac) // TODO: decide on mac passing before merging
	if errl != nil {
		log.Fatalf("%v\n", errl)
	}
	var protocol uint16 = 31337 // TODO: add real protocol
	publicKeyHash := []byte(n.pk.Hash())
	initFrame := l2.NewEthFrame(broadcastAddr, localAddr, protocol, publicKeyHash)
	fmt.Printf("%q: Broadcasting packet.\n", publicKeyHash)
	var err error = frw.WriteFrame(initFrame) // TODO: check errors
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q: Broadcasted packet.\n", publicKeyHash)
	// Process Loop
	go func() {
		for {
			fmt.Printf("%q: Receiving packet.\n", publicKeyHash)
			newInstanceFrame, _ := frw.ReadFrame()
			src := newInstanceFrame.Source()
			dest := newInstanceFrame.Destination()
			fmt.Printf("%q: Received packet from %v.\n", publicKeyHash, src)
			fmt.Printf("%q: Received packet to %v.\n", publicKeyHash, dest)
			if newInstanceFrame.Type() != protocol {
				continue // Throw away if protocols don't match
			}
			if bytes.Equal(src, localAddr) {
				continue // Throw away if from me
			}
			if !(bytes.Equal(dest, localAddr) || bytes.Equal(dest, broadcastAddr)) {
				continue // Throw away if it wasn't to me or the broadcast address
			}
			c <- string(newInstanceFrame.Data())
			if bytes.Equal(dest, broadcastAddr) { // Respond if to broadcast addr
				initFrame := l2.NewEthFrame(src, localAddr, 31337, publicKeyHash) // TODO: add real protocol
				fmt.Printf("%q: Sending response packet %v.\n", publicKeyHash, src)
				var err error = frw.WriteFrame(initFrame) // TODO: check errors
				if err != nil {
					panic(err)
				}
				fmt.Printf("%q: Sent response packet.\n", publicKeyHash)
			}
		}
		close(c)
	}()
	return c, nil // TODO: return channel error?
}
