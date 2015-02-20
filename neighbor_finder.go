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
	var p PublicKey                                      // TODO: pass public key
	publicKeyHash := []byte(p.Hash())
	initFrame := l2.NewEthFrame(broadcastAddr, localAddr, 31337, publicKeyHash) //TODO: add real protocol
	frw.WriteFrame(initFrame)
	// Process Loop
	go func() {
		for {
			newInstanceFrame, _ := frw.ReadFrame()
			src := l2.EthFrame(newInstanceFrame).Source()
			dest := l2.EthFrame(newInstanceFrame).Destination()
			if bytes.Equal(src, localAddr) {
				continue // Throw away if from me
			}
			if !(bytes.Equal(dest, localAddr) || bytes.Equal(dest, broadcastAddr)) {
				continue // Throw away if it wasn't to me or the broadcast address
			}
			c <- string(l2.EthFrame(newInstanceFrame).Data())
			if bytes.Equal(dest, broadcastAddr) { // Respond if to broadcast addr
				var p PublicKey // TODO: pass public key
				publicKeyHash := []byte(p.Hash())
				initFrame := l2.NewEthFrame(src, localAddr, 31337, publicKeyHash) //TODO: add real protocol
				frw.WriteFrame(initFrame)
			}
		}
	}()
	return c
}
