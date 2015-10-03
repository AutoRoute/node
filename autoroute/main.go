package main

import (
	"github.com/AutoRoute/l2"
	"github.com/AutoRoute/node"
	"github.com/AutoRoute/tuntap"

	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
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

type LinkLocalError struct {
	msg   string
	fatal bool
}

func (e LinkLocalError) Error() string { return e.msg }
func (e LinkLocalError) IsFatal() bool { return e.fatal }

func GetLinkLocalAddr(dev net.Interface) (net.IP, error) {
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
		msg := fmt.Sprintf("Couldn't find link-local address on interface %v", dev.Name)
		e := LinkLocalError{msg, false}
		return nil, error(e)
	}

	ll_addr_zone := fmt.Sprintf("%s%%%s", ll_addr.String(), dev.Name)
	_, err = net.ResolveIPAddr("ip6", ll_addr_zone)
	if err != nil {
		e := LinkLocalError{err.Error(), true}
		return nil, error(e)
	}
	return ll_addr, nil
}

func FindNeighbors(dev net.Interface, ll_addr net.IP, key node.PublicKey) <-chan *node.FrameData {
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

func Probe(key node.PrivateKey, n *node.Server, devs []net.Interface) {
	public_key := key.PublicKey()

	// find neighbors of each interface
	for _, dev := range devs {
		if dev.Name == "lo" {
			continue
		}

		ll_addr, err := GetLinkLocalAddr(dev)
		if err != nil {
			e, _ := err.(LinkLocalError)
			if e.IsFatal() {
				log.Fatal(e)
			} else {
				log.Print(e)
				continue
			}
		}

		neighbors := FindNeighbors(dev, ll_addr, public_key)
		go func() {
			for neighbor := range neighbors {
				log.Printf("Neighbour Found %v", neighbor.NodeAddr)
				err := n.Connect(fmt.Sprintf("[%s%%%s]:31337", neighbor.LLAddrStr, dev.Name))
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
		var devs []net.Interface

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
			devs = []net.Interface{*dev}
		}

		log.Printf("Starting to probe on interfaces: ")
		for _, dev := range devs {
			log.Printf("%v", dev.Name)
		}
		go Probe(key, n, devs)
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
