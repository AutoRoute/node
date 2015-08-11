package main

import (
	"github.com/AutoRoute/l2"
	"github.com/AutoRoute/node"

	"code.google.com/p/tuntap"

	"flag"
	"fmt"
	"log"
	"net"
	"strings"
)

var listen = flag.String("listen", "127.0.0.1:34321",
	"The address to listen to incoming connections on")
var nolisten = flag.Bool("nolisten", false, "Disables listening")
var connect = flag.String("connect", "",
	"Comma separated list of addresses to connect to")
var autodiscover = flag.Bool("auto", false,
	"Whether we should try and find neighboring routers")
var tcptun = flag.String("tcptun", "",
	"Address to try and tcp tunnel to")
var keyfile = flag.String("keyfile", "",
	"The keyfile we should check for a key and write our current key to")

func GetLinkLocalAddr(dev net.Interface) (*net.IPAddr, error) {
	dev_addrs, err := dev.Addrs()
	if err != nil {
		return nil, err
	}

	var ll_addr net.IP

	for _, dev_addr := range dev_addrs {
		addr, _, err := net.ParseCIDR(dev_addr.String())
		if err != nil {
			return nil, err
		}

		if addr.IsLinkLocalUnicast() {
			ll_addr = addr
			break
		}
	}

	if ll_addr == nil {
		log.Fatalf("Couldn't find interface link-local addresses %v", dev.Name)
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

func Probe(key node.PrivateKey, n *node.Server) {
	devs, err := net.Interfaces()
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
		go func() {
			for neighbor := range neighbors {
				log.Printf("Neighbour Found %v", neighbor.NodeAddr)
				err := n.Connect(neighbor.LLAddrStr)
				if err != nil {
					log.Printf("Error connecting: %v", err)
					continue
				}
				log.Printf("Connection established to %v", neighbor.NodeAddr)
			}
		}()
	}
}

func main() {
	log.Print("Starting")

	flag.Parse()
	quit := make(chan bool)

	var key node.PrivateKey
	var err error

	if len(*keyfile) > 0 {
		key, err = node.LoadKey(*keyfile)
	} else {
		log.Print("Generating key")
		key, err = node.NewECDSAKey()
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Key is %x", key.PublicKey().Hash())

	n := node.NewServer(key)

	if *autodiscover {
		log.Print("Starting Probing of all interfaces")
		go Probe(key, n)
	}

	if !*nolisten {
		log.Print("Starting Listening")
		err := n.Listen(*listen)
		if err != nil {
			log.Printf("Error listening: %v", err)
		}
	}

	for _, ip := range strings.Split(*connect, ",") {
		if len(ip) == 0 {
			continue
		}
		err := n.Connect(ip)
		if err != nil {
			log.Printf("Error connecting to %s: %v", ip, err)
		}
	}

	if len(*tcptun) > 0 {
		log.Printf("Establishing tcp tunnel to %v", *tcptun)
		i, err := tuntap.Open("tun%d", tuntap.DevTun)
		if err != nil {
			log.Fatal(err)
		}
		dest := ""
		_, err = fmt.Sscanf(*tcptun, "%x", &dest)
		if err != nil {
			log.Fatal(err)
		}
		t := node.NewTCPTunnel(i, n.Node(), node.NodeAddress(dest), 1)
		t = t
		<-quit
	}

	<-quit
}
