package main

import (
	"fmt"
	"github.com/AutoRoute/l2"
	"github.com/AutoRoute/node"
	"log"
	"net"
	"sync"
)

func PrintMessage(cs <-chan string, test string, wg *sync.WaitGroup) {
	defer wg.Done()
	msg := <-cs
	fmt.Printf("%q: Received: %v\n", test, msg)
}

func main() {
	var channels []<-chan string
	interfaces, err := net.Interfaces()

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(interfaces); i++ {
		conn, err := l2.ConnectExistingDevice(interfaces[i].Name)
		if err != nil {
			log.Fatal(err)
		}

		public_key := node.pktest(fmt.Sprintf("test%d", i))
		nf := node.NewNeighborData(public_key)
		out, err := nf.Find(interfaces[i].HardwareAddr, conn)
		if err != nil {
			log.Fatal(err)
		}
		append(channels, out)
	}

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		for j := 0; j < len(channels); j++ {
			wg.Add(1)
			go PrintMessage(channels[i], fmt.Sprintf("test%d", i), &wg)
		}
		wg.Wait()
	}
}
