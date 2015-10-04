package node

import (
	"bytes"
	"log"
	"net"

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
type NeighborFinder struct {
	pk                 PublicKey
	link_local_address net.IP
}

func NewNeighborFinder(pk PublicKey, link_local_address net.IP) NeighborFinder {
	return NeighborFinder{pk, link_local_address}
}

type FrameData struct {
	NodeAddr  NodeAddress
	LLAddrStr string
}

func (n NeighborFinder) handleLink(mac []byte, frw l2.FrameReadWriter, c chan *FrameData) {
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
			continue
		}
		if bytes.Equal(frame.Source(), mac) {
			log.Printf("%q: from ourselves?", n.pk.Hash())
			continue // Throw away if from me
		}
		// If the packet is to us or broadcast, record it.
		if bytes.Equal(frame.Destination(), mac) ||
			bytes.Equal(frame.Destination(), broadcast) {
			data := FrameData{NodeAddress(frame.Data()[:64]), net.IP(frame.Data()[64:]).String()}
			c <- &data
		}
		if !bytes.Equal(frame.Destination(), broadcast) {
			continue
		}
		response := n.BuildResponse(frame, mac, protocol)
		err = frw.WriteFrame(response)
		if err != nil {
			log.Printf("Failure writing to connection %v, %v", frw, err)
			return
		}
	}
}

func (n NeighborFinder) Find(mac []byte, frw l2.FrameReadWriter) (<-chan *FrameData, error) {
	// Send initial packet
	frame_data := append([]byte(n.pk.Hash()), n.link_local_address...)
	frame := l2.NewEthFrame(broadcast, mac, protocol, frame_data)
	err := frw.WriteFrame(frame)
	if err != nil {
		return nil, err
	}

	c := make(chan *FrameData)
	go n.handleLink(mac, frw, c)
	return c, nil
}

func (n NeighborFinder) BuildResponse(frame l2.EthFrame, mac []byte, protocol uint16) l2.EthFrame {
	data := append([]byte(n.pk.Hash()), n.link_local_address...)
	return l2.NewEthFrame(frame.Source(), mac, protocol, data)
}
