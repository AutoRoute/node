package main

import (
	"github.com/AutoRoute/node"
	"github.com/AutoRoute/tuntap"

	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
	"strconv"
	"strings"
)

var listen = flag.String("listen", "127.0.0.1:34321",
	"The address to listen to incoming connections on")
var nolisten = flag.Bool("nolisten", false, "Disables listening")
var connect = flag.String("connect", "",
	"Comma separated list of addresses to connect to")
var autodiscover = flag.Bool("auto", false,
	"Whether we should try and find neighboring routers")
var dev_name = flag.String("interface", "", "Interface to discover on")
var tcptun = flag.String("tcptun", "",
	"Address to try and tcp tunnel to")
var keyfile = flag.String("keyfile", "",
	"The keyfile we should check for a key and write our current key to")

func Probe(key node.PrivateKey, n *node.Server, dev net.Interface, port uint16) {
	if dev.Name == "lo" {
		return
	}

	log.Printf("Probing %q", dev.Name)

	ll_addr, err := node.GetLinkLocalAddr(dev)
	if err != nil {
		log.Printf("Error probing %q: %v", dev.Name, err)
		return
	}

	neighbors := node.FindNeighbors(dev, ll_addr, key.PublicKey(), port)
	for neighbor := range neighbors {
		log.Printf("Neighbour Found %x", neighbor.NodeAddr)
		err := n.Connect(fmt.Sprintf("[%s%%%s]:%v", neighbor.LLAddrStr, dev.Name, neighbor.Port))
		if err != nil {
			log.Printf("Error connecting: %v", err)
			return
		}
		log.Printf("Connection established to %x", neighbor.NodeAddr)
	}
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

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
		devs := make([]net.Interface, 0)

		if *dev_name == "" {
			devs, err = net.Interfaces()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			dev, err := net.InterfaceByName(*dev_name)
			if err != nil {
				log.Fatal(err)
			}
			devs = append(devs, *dev)
		}

		parsed_listen_addr := strings.Split(*listen, ":")
		port64, err := strconv.ParseUint(parsed_listen_addr[len(parsed_listen_addr)-1], 10, 16)
		port := uint16(port64)
		if err != nil {
			log.Fatal(err)
		}
		port = uint16(port)
		for _, dev := range devs {
			go Probe(key, n, dev, port)
		}
	}

	if !*nolisten {
		log.Print("Starting to listen")
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
