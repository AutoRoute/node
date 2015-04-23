package node

import (
	"golang.org/x/crypto/ssh"
)

struct SSHConnection {
  addresses []NodeAddress
  sessions  []*ssh.Session
}

func (s SSHConnection) SendMap(ReachabilityMap) error {
}

func (s SSHConnection) ReachabilityMaps() <-chan ReachabilityMap {
}

func (s SSHConnection) SendReceipts() <-chan PacketReceipt {
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
