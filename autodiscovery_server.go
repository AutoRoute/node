package node

import (
	"errors"
	"log"
	"net"

	"github.com/AutoRoute/l2"

	"github.com/AutoRoute/node/internal"
)

func GetLinkLocalAddr(dev net.Interface) (net.IP, error) {
	dev_addrs, err := dev.Addrs()
	if err != nil {
		return nil, err
	}

	for _, dev_addr := range dev_addrs {
		addr, _, err := net.ParseCIDR(dev_addr.String())
		if err != nil {
			return nil, err
		}

		if addr.IsLinkLocalUnicast() {
			return addr, nil
		}
	}
	return nil, errors.New("Unable to find Link local address")
}

func FindNeighbors(dev net.Interface, ll_addr net.IP, key node.PublicKey, port uint16) <-chan *FrameData {
	conn, err := l2.ConnectExistingDevice(dev.Name)
	if err != nil {
		log.Fatal(err)
	}
	nf := NewNeighborFinder(key, ll_addr, port)
	channel, err := nf.Find(dev.HardwareAddr, conn)
	if err != nil {
		log.Fatal(err)
	}
	return channel
}
