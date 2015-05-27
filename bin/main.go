package main

import (
	"fmt"
	"github.com/AutoRoute/l2"
	"github.com/AutoRoute/node"
	"log"
	"net"
	"sync"
)

func FindNeighbors(interfaces []net.Interface, public_key node.PublicKey) []<-chan node.NodeAddress {
	var channels []<-chan node.NodeAddress
	for i := 0; i < len(interfaces); i++ {
		conn, err := l2.ConnectExistingDevice(interfaces[i].Name)
		if err != nil {
			log.Fatal(err)
		}

		nf := node.NewNeighborData(public_key)
		out, err := nf.Find(interfaces[i].HardwareAddr, conn)
		if err != nil {
			log.Fatal(err)
		}
		channels = append(channels, out)
	}
	return channels
}

func ReadResponses(channels []<-chan node.NodeAddress) []node.NodeAddress {
	var wg sync.WaitGroup
	var addresses []node.NodeAddress
	for i := 0; i < 2; i++ {
		for j := 0; j < len(channels); j++ {
			wg.Add(1)
			go func(cs <-chan node.NodeAddress, test string, wg *sync.WaitGroup) {
				defer wg.Done()
				msg := <-cs
				addresses = append(addresses, msg)
				fmt.Printf("%q: Received: %v\n", test, msg)
			}(channels[i], fmt.Sprintf("test%d", i), &wg)
		}
		wg.Wait()
	}
	return addresses
}

func CloseConnections(connections []*node.SSHConnection) error {
	for i := 0; i < len(connections); i++ {
		err := connections[i].Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	interfaces, err := net.Interfaces()

	if err != nil {
		log.Fatal(err)
	}

	k, err := node.NewECDSAKey()

	if err != nil {
		log.Fatal(err)
	}

	public_key := k.PublicKey()

	// find neighbors of each interface
	channels := FindNeighbors(interfaces, public_key)

	// print responses and return addresses (public key hashes)
	addresses := ReadResponses(channels)

	// establish ssh connections
	connections := node.EstablishSSH(addresses, k)
	defer CloseConnections(connections)
}
