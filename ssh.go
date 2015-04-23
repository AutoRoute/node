package node

import (
	"golang.org/x/crypto/ssh"
)

type SSHConnection struct {
	address NodeAddress
	session *ssh.Session
}

func (s SSHConnection) SendMap(ReachabilityMap) error {
}

func (s SSHConnection) ReachabilityMaps() <-chan ReachabilityMap {
}

func (s SSHConnection) SendReceipts() <-chan PacketReceipt {
}

func (s SSHConnection) PacketReceipts() <-chan PacketReceipt {
}

func (s SSHConnection) SendPayment(Payment) error {
}

func (s SSHConnection) Payments() <-chan Payment {
}

func (s SSHConnectoin) SendPacket(Packet) error {
}

func (s SSHConnection) Packets() <-chan Packet {
}

func (s SSHConnection) Key() PublicKey {
}

func (s SSHConnection) Close() error {
	for i := 0; i < len(ssessions); i++ {
		s.sessions[i].Close()
	}
}

func EstablishSSH(addresses []node.NodeAddress, key NodeAddress) []*SSHConnection {
	username := "node-username"
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
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
