package main

import (
	"github.com/AutoRoute/l2"
	"github.com/AutoRoute/node"

	"log"
	"net"
)

func FindNeighbors(dev net.Interface, key node.PublicKey) <-chan node.NodeAddress {
	conn, err := l2.ConnectExistingDevice(dev.Name)
	if err != nil {
		log.Fatal(err)
	}

	nf := node.NewNeighborData(key)
	channel, err := nf.Find(dev.HardwareAddr, conn)
	if err != nil {
		log.Fatal(err)
	}
	return channel
}

func main() {
	devs, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	key, err := node.NewECDSAKey()
	if err != nil {
		log.Fatal(err)
	}

	public_key := key.PublicKey()

	// find neighbors of each interface
	for _, dev := range devs {
		neighbours := FindNeighbors(dev, public_key)
		for addr := range neighbours {
			log.Printf("Neighbour Found %v", string(addr))
			c, err := net.Dial("tcp", string(addr))
			if err != nil {
				log.Printf("Error connecting %v", err)
			}
			connection, err := node.EstablishSSH(c, string(addr), key)
			if err != nil {
				log.Printf("Error connecting: %v", err)
			}
			log.Printf("Connection established to %v %v", addr, connection)
		}
	}
}
