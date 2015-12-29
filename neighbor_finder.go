package node

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"

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
	pk                 internal.PublicKey
	Link_local_address net.IP
	port               uint16
}

func NewNeighborFinder(pk internal.PublicKey, Link_local_address net.IP, port uint16) NeighborFinder {
	return NeighborFinder{pk, Link_local_address, port}
}

type FrameData struct {
	fullNodeAddr  types.NodeAddress
	LLAddrStr string
	Port      uint16
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
			port_buf := bytes.NewBuffer(frame.Data()[80:])
			var port uint16
			err := binary.Read(port_buf, binary.LittleEndian, &port)
			if err != nil {
				log.Fatal(err)
			}
			data := FrameData{types.NodeAddress(frame.Data()[:64]), net.IP(frame.Data()[64:80]).String(), port}
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
	var port_buf bytes.Buffer
	err := binary.Write(&port_buf, binary.LittleEndian, n.port)
	if err != nil {
		log.Fatal(err)
	}
	frame_data := append([]byte(n.pk.Hash()), n.Link_local_address...)
	frame_data = append(frame_data, port_buf.Bytes()...)
	frame := l2.NewEthFrame(broadcast, mac, protocol, frame_data)
	err = frw.WriteFrame(frame)
	if err != nil {
		return nil, err
	}

	c := make(chan *FrameData)
	go n.handleLink(mac, frw, c)
	return c, nil
}

func (n NeighborFinder) BuildResponse(frame l2.EthFrame, mac []byte, protocol uint16) l2.EthFrame {
	var port_buf bytes.Buffer
	err := binary.Write(&port_buf, binary.LittleEndian, n.port)
	if err != nil {
		log.Fatal(err)
	}
	data := append([]byte(n.pk.Hash()), n.Link_local_address...)
	data = append(data, port_buf.Bytes()...)
	return l2.NewEthFrame(frame.Source(), mac, protocol, data)
}

func Probe(key Key, n *Server, dev net.Interface, port uint16) {
	if dev.Name == "lo" {
		return
	}

	log.Printf("Probing %q", dev.Name)

	ll_addr, err := GetLinkLocalAddr(dev)
	if err != nil {
		log.Printf("Error probing %q: %v", dev.Name, err)
		return
	}

	neighbors := FindNeighbors(dev, ll_addr, key.k.PublicKey(), port)
	for neighbor := range neighbors {
		log.Printf("Neighbour Found %x", neighbor.fullNodeAddr)
		err := n.Connect(fmt.Sprintf("[%s%%%s]:%v", neighbor.LLAddrStr, dev.Name, neighbor.Port))
		if err != nil {
			log.Printf("Error connecting: %v", err)
			return
		}
		log.Printf("Connection established to %x", neighbor.fullNodeAddr)
	}
}
