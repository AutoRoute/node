package main

import (
	"github.com/AutoRoute/l2"
	"github.com/AutoRoute/node"

	"flag"
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

func Prove(key node.PrivateKey, node node.Node) {
	devs, err := net.Interfaces()
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

func Connect(addr string, key node.PrivateKey) (*node.SSHConnection, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return node.EstablishSSH(c, addr, key)
}

func Listen(key node.PrivateKey, node node.Node) {
	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}

	l := node.ListenSSH(ln, key)
	for c := range l.Connections() {
		log.Printf("Incoming connection: %v", c)
		node.AddConnection(c)
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

	node := NewNode(key, time.Tick(time.Second), time.Tick(time.Second))

	if *autodiscover {
		log.Print("Starting Probing of all interfaces")
		go Probe(key, node)
	}

	if !*nolisten {
		log.Print("Starting Listening")
		go Listen(key, node)
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
		node.AddConnection(c)
	}

	<-quit
}
