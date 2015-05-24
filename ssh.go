package node

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

type SSHConnection struct {
	address NodeAddress
	session *ssh.Session
}

func (s SSHConnection) SendMap(ReachabilityMap) error {
	return nil
}

func (s SSHConnection) ReachabilityMaps() <-chan ReachabilityMap {
	return nil
}

func (s SSHConnection) SendReceipts() <-chan PacketReceipt {
	return nil
}

func (s SSHConnection) PacketReceipts() <-chan PacketReceipt {
	return nil
}

func (s SSHConnection) SendPayment(Payment) error {
	return nil
}

func (s SSHConnection) Payments() <-chan Payment {
	return nil
}

func (s SSHConnection) SendPacket(Packet) error {
	return nil
}

func (s SSHConnection) Packets() <-chan Packet {
	return nil
}

func (s SSHConnection) Key() PublicKey {
	return nil
}

func (s SSHConnection) Close() error {
	err := s.session.Close()
	return err
}

func EstablishSSH(addresses []NodeAddress, key PrivateKey) []*SSHConnection {
	username := "node-username"
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		panic("Failed to create signer from key")
	}
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	var connections []*SSHConnection
	for i := 0; i < len(addresses); i++ {
		client, err := ssh.Dial("tcp", fmt.Sprintf(string(addresses[i]), ":22"), config)
		if err != nil {
			panic("Failed to dial: " + err.Error())
		}
		session, err := client.NewSession()
		if err != nil {
			panic("Failed to create session: " + err.Error())
		}
		connection := SSHConnection{address: addresses[i], session: session}
		connections = append(connections, &connection)
	}
	return connections
}
