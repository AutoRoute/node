package main

import (
	"github.com/AutoRoute/l2"
	"github.com/AutoRoute/node"

	"fmt"
	"log"
	"net"
)

func GetLinkLocalAddr(dev net.Interface) (*net.IPAddr, error) {
	cidr_ll_addr, err := dev.Addrs()
	if err != nil {
		return nil, err
	}

	ll_addr, _, err := net.ParseCIDR(cidr_ll_addr[1].String())
	if err != nil {
		return nil, err
	}

	ll_addr_zone := fmt.Sprintf("%s%%%s", ll_addr.String(), dev.Name)
	resolved_ll_addr, err := net.ResolveIPAddr("ip6", ll_addr_zone)
	if err != nil {
		return nil, err
	}
	return resolved_ll_addr, nil
}

func FindNeighbors(dev net.Interface, ll_addr *net.IPAddr, key node.PublicKey) <-chan *node.FrameData {
	conn, err := l2.ConnectExistingDevice(dev.Name)
	if err != nil {
		log.Fatal(err)
	}

	nf := node.NewNeighborData(key, ll_addr)
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
		if dev.Name == "lo" {
			continue
		}

		ll_addr, err := GetLinkLocalAddr(dev)
		if err != nil {
			log.Fatal(err)
		}

		neighbors := FindNeighbors(dev, ll_addr, public_key)
		for neighbor := range neighbors {
			log.Printf("Neighbour Found %v", neighbor.NodeAddr)
			connection, err := node.EstablishSSH(neighbor.LLAddrStr, key)
			if err != nil {
				log.Printf("Error connecting: %v", err)
			}
			log.Printf("Connection established to %v %v", neighbor.NodeAddr, connection)
		}
	}
}
