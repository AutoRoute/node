package main

import (
	"github.com/AutoRoute/l2"
	"github.com/AutoRoute/node"

	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

var listen = flag.String("listen", "127.0.0.1:34321",
	"The address to listen to incoming connections on")
var nolisten = flag.Bool("nolisten", false, "Disables listening")
var connect = flag.String("connect", "",
	"Comma separated list of addresses to connect to")
var autodiscover = flag.Bool("auto", false,
	"Whether we should try and find neighboring routers")

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

func Probe(key node.PrivateKey, n node.Node) {
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
		for neighbor := range neighbors {
			log.Printf("Neighbour Found %v", neighbor.NodeAddr)
			c, err := net.Dial("tcp", neighbor.LLAddrStr)
			if err != nil {
				log.Printf("Error connecting %v", err)
			}
			connection, err := node.EstablishSSH(c, neighbor.LLAddrStr, key)
			if err != nil {
				log.Printf("Error connecting: %v", err)
			}
			log.Printf("Connection established to %v %v", neighbor.NodeAddr, connection)
			n.AddConnection(connection)
		}
	}
}

func Connect(addr string, key node.PrivateKey) (*node.SSHConnection, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return node.EstablishSSH(c, addr, key)
}

func Listen(key node.PrivateKey, n node.Node) {
	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}

	l := node.ListenSSH(ln, key)
	for c := range l.Connections() {
		log.Printf("Incoming connection: %v", c)
		n.AddConnection(c)
	}
	log.Printf("Closing error: %v", l.Error())
}

func main() {
	log.Print("Starting")

	flag.Parse()
	quit := make(chan bool)

	log.Print("Generating key")
	key, err := node.NewECDSAKey()
	if err != nil {
		log.Fatal(err)
	}

	n := node.NewNode(key, time.Tick(time.Second), time.Tick(time.Second))

	if *autodiscover {
		log.Print("Starting Probing of all interfaces")
		go Probe(key, n)
	}

	if !*nolisten {
		log.Print("Starting Listening")
		go Listen(key, n)
	}

	for _, ip := range strings.Split(*connect, ",") {
		if len(ip) == 0 {
			continue
		}
		c, err := Connect(ip, key)
		if err != nil {
			log.Printf("Error connecting to %s: %v", ip, err)
		} else {
			log.Printf("Outgoing connection: %v", c)
		}
		n.AddConnection(c)
	}

	<-quit
}
