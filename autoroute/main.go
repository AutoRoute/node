// This binary is the canonical way to use autoroute. It supports most use cases and should
// be able to provide most needs. Ideally the binary should merely be a wrapper around the core library
// and should be the canonical example of how to use it. There should be nothing which this binary can
// do which isn't possible in the core library.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/AutoRoute/node"
	"github.com/AutoRoute/node/types"
	"github.com/AutoRoute/tuntap"
)

var listen = flag.String("listen", "[::]:34321",
	"The address to listen to incoming connections on")
var connect = flag.String("connect", "",
	"Comma separated list of addresses to connect to")
var autodiscover = flag.Bool("auto", false,
	"Whether we should try and find neighboring routers")
var dev_names = flag.String("devs", "",
	"Comma separated list of interfaces to discover on")
var keyfile = flag.String("keyfile", "",
	"The keyfile we should check for a key and write our current key to")
var btc_host = flag.String("btc_host", "localhost:8333",
	"The bitcoin daemon we should connect to")
var btc_user = flag.String("btc_user", "user",
	"The bitcoin daemon username")
var btc_pass = flag.String("btc_pass", "password",
	"The bitcoin daemon password")
var fake_money = flag.Bool("fake_money", false, "Enables a money system which is purely fake")
var status = flag.String("status", "[::1]:12345", "The port to expose status information on")
var unix = flag.String("unix", "", "The path to accept / receive packets as unix packets from")
var tcptun = flag.String("tcptun", "", "Address to try and tcp tunnel to")
var tcptunserve = flag.Bool("tcptunserve", false, "Enables this node to be an exit node")
var tcpaddress = flag.String("tcp_address", "", "IP address to assign to the tcp tunnel")

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Print(os.Args)
	flag.Parse()

	// Capture all signals to the quit channel
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)

	// Figure out and load what key we are using for our identity
	var key node.Key
	var err error
	if len(*keyfile) > 0 {
		key, err = node.LoadKey(*keyfile)
		if err != nil {
			log.Fatalf("Error loading key: %v", err)
		}
	} else {
		log.Print("Generating key")
		key, err = node.NewKey()
		if err != nil {
			log.Fatalf("Error Generating Key: %v", err)
		}
	}
	log.Printf("Key is %x", key)

	// Connect to our money server
	money := node.FakeMoney()
	if !*fake_money {
		log.Print("Connecting to bitcoin daemon")
		rpc, err := node.NewRPCMoney(*btc_host, *btc_user, *btc_pass)
		if err != nil {
			log.Fatalf("Error connection to bitcoin daemon: %v", err)
		}
		money = rpc
	}
	log.Printf("Connected")
	var packet_logger node.Logger
	n := node.NewServer(key, money, log.New(os.Stderr, "", log.LstdFlags), packet_logger)

	log.Printf("Starting to listen on %s", *listen)
	err = n.Listen(*listen)
	if err != nil {
		log.Printf("Error listening: %v", err)
	}

	if *autodiscover {
		devs := make([]net.Interface, 0)

		if *dev_names == "" {
			devs, err = net.Interfaces()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			for _, dev_name := range strings.Split(*dev_names, ",") {
				dev, err := net.InterfaceByName(dev_name)
				if err != nil {
					log.Fatal(err)
				}
				devs = append(devs, *dev)
			}
		}

		for _, dev := range devs {
			go func(dev net.Interface) {
				err := n.Probe(dev)
				if err != nil {
					log.Fatal(err)
				}
			}(dev)
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

	go func() {
		log.Fatal(http.ListenAndServe(*status, nil))
	}()

	if *tcptunserve {
		log.Printf("Starting tcp tunnel server")
		log.Printf("Establishing tcp tunnel to %v", *tcptun)
		i, err := tuntap.Open("tun%d", tuntap.DevTun)
		if err != nil {
			log.Fatal(err)
		}
		tunserver := node.NewTCPTunServer(n.Node(), i, 10000)
		tunserver.Listen()
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
		t := node.NewTCPTunClient(n.Node(), i, types.NodeAddress(dest), 10000, i.Name())
		defer t.Close()

		if len(*tcpaddress) > 0 {
			ip, _, err := net.ParseCIDR(*tcpaddress)
			if err != nil {
				log.Fatal(err)
			}
			err = exec.Command("ip", "link", "set", "dev", i.Name(), "up").Run()
			if err != nil {
				log.Fatal(err)
			}
			err = exec.Command("ip", "addr", "add", *tcpaddress, "dev", i.Name()).Run()
			if err != nil {
				log.Fatal(err)
			}
			err = exec.Command("ip", "route", "add", "0/1", "via", ip.String(), "dev", i.Name()).Run()
			if err != nil {
				log.Fatal(err)
			}
			err = exec.Command("ip", "route", "add", "128/1", "via", ip.String(), "dev", i.Name()).Run()
			if err != nil {
				log.Fatal(err)
			}

		}
		m := <-quit
		log.Print(m)
		return
	}

	if len(*unix) > 0 {
		log.Printf("Establishing unix interface %s", *unix)
		c, err := node.NewUnixSocket(*unix, n.Node())
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(*unix)
		defer c.Close()
		<-quit
		return
	}

	<-quit
}
