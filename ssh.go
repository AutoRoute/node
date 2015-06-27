package node

import (
	"net"

	"golang.org/x/crypto/ssh"
)

type SSHConnection struct {
	conn  ssh.Conn
	chans <-chan ssh.NewChannel
	reqs  <-chan *ssh.Request
}

func (s SSHConnection) SendMap(ReachabilityMap) error {
	return nil
}

func (s SSHConnection) ReachabilityMaps() <-chan ReachabilityMap {
	return nil
}

func (s SSHConnection) SendReceipts(PacketReceipt) error {
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
	return PublicKey{}
}

func (s SSHConnection) Close() error {
	err := s.conn.Close()
	return err
}

type SSHListener struct {
	err error
	c   chan *SSHConnection
}

func (l *SSHListener) Error() error {
	return l.err
}

func (l *SSHListener) Connections() <-chan *SSHConnection {
	return l.c
}

func ListenSSH(address string, key PrivateKey) SSHListener {
	l := SSHListener{nil, make(chan *SSHConnection)}
	l.listen(address, key)
	return l
}

func (l *SSHListener) error(err error) {
	l.err = err
	close(l.c)
	return
}

func (l *SSHListener) listen(address string, key PrivateKey) {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{}, nil
		},
	}
	signer, err := ssh.NewSignerFromKey(key.k)
	if err != nil {
		l.error(err)
		return
	}
	config.AddHostKey(signer)
	s, err := net.Listen("tcp", address)
	if err != nil {
		l.error(err)
		return
	}
	go func() {
		for {
			c, err := s.Accept()
			if err != nil {
				l.error(err)
				return
			}

			server, chans, reqs, err := ssh.NewServerConn(c, config)
			if err != nil {
				l.error(err)
				return
			}
			l.c <- &SSHConnection{server, chans, reqs}
		}
	}()
}

func EstablishSSH(address string, key PrivateKey) (*SSHConnection, error) {
	username := string(key.PublicKey().Hash())
	signer, err := ssh.NewSignerFromKey(key.k)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	c, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	client, chans, reqs, err := ssh.NewClientConn(c, address, config)
	if err != nil {
		return nil, err
	}
	return &SSHConnection{client, chans, reqs}, nil
}
