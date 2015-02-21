package node

import (
	"bytes"
	"github.com/AutoRoute/l2"
)

// The layer two protocol takes a layer two device and returns the hash of the
// Public Key of all neighbors it can find.
type NeighborFinder interface {
	Find(l2.FrameReadWriter) <-chan string
}

type layer2 struct{}

func (nf layer2) Find(frw l2.FrameReadWriter) <-chan string {
	c := make(chan string)
	// Broadcast Hash
	broadcastAddr := l2.MacToBytesOrDie("ff:ff:ff:ff:ff:ff")
	localAddr := l2.MacToBytesOrDie("aa:bb:cc:dd:ee:00") // TODO: pass own mac address
	var protocol uint16 = 31337                          // TODO: add real protocol
	var p PublicKey                                      // TODO: pass public key
	publicKeyHash := []byte(p.Hash())
	initFrame := l2.NewEthFrame(broadcastAddr, localAddr, protocol, publicKeyHash)
	_ = frw.WriteFrame(initFrame) // TODO: check errors
	// Process Loop
	go func() {
		for {
			newInstanceFrame, _ := frw.ReadFrame()
			src := newInstanceFrame.Source()
			dest := newInstanceFrame.Destination()
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
				var p PublicKey // TODO: pass public key
				publicKeyHash := []byte(p.Hash())
				initFrame := l2.NewEthFrame(src, localAddr, 31337, publicKeyHash) // TODO: add real protocol
				_ = frw.WriteFrame(initFrame)                                     // TODO: check errors
			}
		}
	}()
	return c
}
