package node

import (
	"bytes"
	"log"

	"github.com/AutoRoute/l2"
)

const protocol = 6042

var broadcast []byte

func init() {
	var err error
	broadcast, err = l2.MacToBytes("ff:ff:ff:ff:ff:ff")
	if err != nil {
		log.Fatal(err)
	}
}

// The layer two protocol takes a layer two device and returns the hash of the
// Public Key of all neighbors it can find.
type NeighborFinder interface {
	Find(l2.FrameReadWriter) <-chan NodeAddress
}

type NeighborData struct {
	pk PublicKey
}

func NewNeighborData(pk PublicKey) NeighborData {
	return NeighborData{pk}
}

func (n NeighborData) handleLink(mac []byte, frw l2.FrameReadWriter, c chan NodeAddress) {
	// Handle received packets
	defer close(c)
	for {
		frame, err := frw.ReadFrame()
		if err != nil {
			log.Printf("Failure reading from connection %v, %v", frw, err)
			return
		}
		// Throw away if protocols don't match
		if frame.Type() != protocol {
			log.Printf("%q: bad protocol", n.pk.Hash())
			continue
		}
		if bytes.Equal(frame.Source(), mac) {
			log.Printf("%q: from ourselves?", n.pk.Hash())
			continue // Throw away if from me
		}
		// If the packet is to us or broadcast, record it.
		if bytes.Equal(frame.Destination(), mac) ||
			bytes.Equal(frame.Destination(), broadcast) {
			c <- NodeAddress(string(frame.Data()))
		}
		if !bytes.Equal(frame.Destination(), broadcast) {
			continue
		}
		response := l2.NewEthFrame(frame.Source(), mac, protocol, []byte(n.pk.Hash()))
		err = frw.WriteFrame(response)
		if err != nil {
			log.Printf("Failure writing to connection %v, %v", frw, err)
			return
		}
	}
}

func (n NeighborData) Find(mac []byte, frw l2.FrameReadWriter) (<-chan NodeAddress, error) {
	// Send initial packet
	frame := l2.NewEthFrame(broadcast, mac, protocol, []byte(n.pk.Hash()))
	err := frw.WriteFrame(frame)
	if err != nil {
		return nil, err
	}

	c := make(chan NodeAddress)
	go n.handleLink(mac, frw, c)
	return c, nil
}
