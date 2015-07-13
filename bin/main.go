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

func FindNeighbors(dev net.Interface, ll_addr *net.IPAddr, key node.PublicKey) <-chan node.NodeAddress {
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
		ll_addr, err := GetLinkLocalAddr(dev)
		if err != nil {
			log.Fatal(err)
		}

		neighbours := FindNeighbors(dev, ll_addr, public_key)
		for addr := range neighbours {
			log.Printf("Neighbour Found %v", addr)
			connection, err := node.EstablishSSH("Dummy address", key)
			if err != nil {
				log.Printf("Error connecting: %v", err)
			}
			log.Printf("Connection established to %v %v", addr, connection)
		}
	}
}
