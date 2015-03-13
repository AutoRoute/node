package node

import (
	"bytes"
	"fmt"
	"github.com/AutoRoute/l2"
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
		return c, errb
	}
	localAddr, errl := l2.MacToBytes(mac) // TODO: decide on mac passing before merging
	if errl != nil {
		return c, errl
	}
	var protocol uint16 = 31337 // TODO: add real protocol
	publicKeyHash := []byte(n.pk.Hash())
	initFrame := l2.NewEthFrame(broadcastAddr, localAddr, protocol, publicKeyHash)
	fmt.Println("Broadcasting packet.")
	var err error = frw.WriteFrame(initFrame) // TODO: check errors
	if err != nil {
		panic(err)
	}
	fmt.Println("Broadcasted packet.")
	// Process Loop
	go func() {
		for {
			fmt.Println("Receiving packet.")
			newInstanceFrame, _ := frw.ReadFrame()
			src := newInstanceFrame.Source()
			dest := newInstanceFrame.Destination()
			fmt.Printf("Received packet from %v.\n", src)
			fmt.Printf("Received packet to %v.\n", dest)
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
				fmt.Printf("Sending response packet %v.\n", src)
				var err error = frw.WriteFrame(initFrame) // TODO: check errors
				if err != nil {
					panic(err)
				}
				fmt.Println("Sent response packet.")
			}
		}
	}()
	return c, nil // TODO: return channel error?
}
