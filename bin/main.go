package main

import (
	"fmt"
	"github.com/AutoRoute/l2"
	"github.com/AutoRoute/node"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os/user"
	"sync"
)

func FindNeighbors(interfaces []net.Interface) []<-chan node.NodeAddress {
	var channels []<-chan node.NodeAddress
	for i := 0; i < len(interfaces); i++ {
		conn, err := l2.ConnectExistingDevice(interfaces[i].Name)
		if err != nil {
			log.Fatal(err)
		}

		k, err := node.NewECDSAKey()
		public_key := k.PublicKey()
		if err != nil {
			log.Fatal(err)
		}

		nf := node.NewNeighborData(public_key)
		out, err := nf.Find(interfaces[i].HardwareAddr, conn)
		if err != nil {
			log.Fatal(err)
		}
		channels = append(channels, out)
	}
	return channels
}

func ReadResponses(channels []<-chan node.NodeAddress) []node.NodeAddress {
	var wg sync.WaitGroup
	var addresses []node.NodeAddress
	for i := 0; i < 2; i++ {
		for j := 0; j < len(channels); j++ {
			wg.Add(1)
			go func(cs <-chan node.NodeAddress, test string, wg *sync.WaitGroup) {
				defer wg.Done()
				msg := <-cs
				addresses = append(addresses, msg)
				fmt.Printf("%q: Received: %v\n", test, msg)
			}(channels[i], fmt.Sprintf("test%d", i), &wg)
		}
		wg.Wait()
	}
	return addresses
}

func getKeyFile() (key ssh.Signer, err error) {
	usr, _ := user.Current()
	file := usr.HomeDir + "/.ssh/id_rsa"
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return
	}
	return
}

func EstablishSSH(addresses []node.NodeAddress) []*ssh.Session {
	key, err := getKeyFile()
	if err != nil {
		panic(err)
	}
	username := "node-username"
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	var sessions []*ssh.Session
	for i := 0; i < len(addresses); i++ {
		client, err := ssh.Dial("tcp", fmt.Sprintf(string(addresses[i]), ":22"), config)
		if err != nil {
			panic("Failed to dial: " + err.Error())
		}
		session, err := client.NewSession()
		if err != nil {
			panic("Failed to create session: " + err.Error())
		}
		sessions = append(sessions, session)
	}
	return sessions
}

func CloseSessions(sessions []*ssh.Session) {
	for i := 0; i < len(sessions); i++ {
		sessions[i].Close()
	}
}

func main() {
	interfaces, err := net.Interfaces()

	if err != nil {
		log.Fatal(err)
	}

	// find neighbors of each interface
	channels := FindNeighbors(interfaces)

	// print responses and return addresses (public key hashes)
	addresses := ReadResponses(channels)

	// establish ssh connections
	sessions := EstablishSSH(addresses)
	defer CloseSessions(sessions)
}
