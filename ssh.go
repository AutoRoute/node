package node

import (
	"encoding/json"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type SSHChannel struct {
	c    ssh.Channel
	reqs <-chan *ssh.Request
}

type SSHConnection struct {
	conn  ssh.Conn
	chans <-chan ssh.NewChannel
	reqs  <-chan *ssh.Request
	c     map[string]*SSHChannel
	e     map[string]*json.Encoder
	d     map[string]*json.Decoder
}

func NewSSHConnection(conn ssh.Conn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) *SSHConnection {
	s := &SSHConnection{conn, chans, reqs, make(map[string]*SSHChannel), make(map[string]*json.Encoder), make(map[string]*json.Decoder)}
	// Set up various session types we want.
	s.c["reachability"] = nil
	s.c["receipt"] = nil
	s.c["payment"] = nil
	s.c["packet"] = nil
	return s
}

func (s SSHConnection) Connect() error {
	err := s.connect("reachability")
	if err != nil {
		s.conn.Close()
		return err
	}
	err = s.connect("receipt")
	if err != nil {
		s.conn.Close()
		return err
	}
	err = s.connect("payment")
	if err != nil {
		s.conn.Close()
		return err
	}
	err = s.connect("packet")
	if err != nil {
		s.conn.Close()
		return err
	}
	return nil
}

func (s SSHConnection) Listen() {
	for nc := range s.chans {
		if _, ok := s.c[nc.ChannelType()]; !ok {
			nc.Reject(ssh.UnknownChannelType, "Unknown channel type")
			continue
		}
		if s.c[nc.ChannelType()] == nil {
			c, r, err := nc.Accept()
			if err != nil {
				log.Printf("Error accepting channel request: %v", err)
				continue
			}
			s.c[nc.ChannelType()] = &SSHChannel{c, r}
			s.e[nc.ChannelType()] = json.NewEncoder(c)
			s.d[nc.ChannelType()] = json.NewDecoder(c)
		} else {
			nc.Reject(ssh.ConnectionFailed, "Connection already established")
		}
	}
}

func (s SSHConnection) connect(name string) error {
	c, r, err := s.conn.OpenChannel(name, nil)
	if err != nil {
		return err
	}
	s.c[name] = &SSHChannel{c, r}
	s.e[name] = json.NewEncoder(c)
	s.d[name] = json.NewDecoder(c)
	return nil
}

func (s SSHConnection) SendMap(m ReachabilityMap) error {
	return s.e["reachability"].Encode(m)
}

func (s SSHConnection) ReachabilityMaps() <-chan ReachabilityMap {
	c := make(chan ReachabilityMap)
	go func() {
		var v ReachabilityMap
		err := s.d["reachability"].Decode(&v)
		if err != nil {
			close(c)
		} else {
			c <- v
		}
	}()
	return c
}

func (s SSHConnection) SendReceipts(r PacketReceipt) error {
	return s.e["receipt"].Encode(r)
}

func (s SSHConnection) PacketReceipts() <-chan PacketReceipt {
	c := make(chan PacketReceipt)
	go func() {
		var v PacketReceipt
		err := s.d["receipt"].Decode(&v)
		if err != nil {
			close(c)
		} else {
			c <- v
		}
	}()
	return c
}

func (s SSHConnection) SendPayment(p Payment) error {
	return s.e["payment"].Encode(p)
}

func (s SSHConnection) Payments() <-chan Payment {
	c := make(chan Payment)
	go func() {
		var v Payment
		err := s.d["payment"].Decode(&v)
		if err != nil {
			close(c)
		} else {
			c <- v
		}
	}()
	return c
}

func (s SSHConnection) SendPacket(p Packet) error {
	return s.e["packet"].Encode(p)
}

func (s SSHConnection) Packets() <-chan Packet {
	c := make(chan Packet)
	go func() {
		var v Packet
		err := s.d["packet"].Decode(&v)
		if err != nil {
			close(c)
		} else {
			c <- v
		}
	}()
	return c
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
			sc := NewSSHConnection(server, chans, reqs)
			go sc.Listen()
			l.c <- sc
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
	sc := NewSSHConnection(client, chans, reqs)
	err = sc.Connect()
	return sc, err
}
