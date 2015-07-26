package node

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

// Represents a single ssh channel, which is being written to by a Encoder / Decoder.
type SSHChannel struct {
	c    ssh.Channel
	reqs <-chan *ssh.Request
}

// Represents an active SSH connection with another host. Contains multiple
// channels passing various message types. Satisfies the Connection interface.
type SSHConnection struct {
	conn  ssh.Conn
	chans <-chan ssh.NewChannel
	reqs  <-chan *ssh.Request
	c     map[string]*SSHChannel
	e     map[string]*json.Encoder
	el    map[string]*sync.Mutex
	d     map[string]*json.Decoder
	dl    map[string]*sync.Mutex
	key   PublicKey
	// Channel to allow blocking until ssh channels are established
	sync chan bool
	lock *sync.Mutex
}

// Constructs a new SSHConnection given the various items returned by the /x/c/ssh library.
// Does *not* call listen() or connect() on the SSHConnection, which is required
// to establish the various required channels.
func NewSSHConnection(conn ssh.Conn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request, key PrivateKey) *SSHConnection {
	s := &SSHConnection{conn,
		chans,
		reqs,
		make(map[string]*SSHChannel),
		make(map[string]*json.Encoder),
		make(map[string]*sync.Mutex),
		make(map[string]*json.Decoder),
		make(map[string]*sync.Mutex),
		PublicKey{},
		make(chan bool),
		&sync.Mutex{},
	}
	// Set up various session types we want.
	for _, k := range []string{"reachability", "receipt", "payment", "packet"} {
		s.c[k] = nil
		s.el[k] = &sync.Mutex{}
		s.dl[k] = &sync.Mutex{}
	}
	go s.sendKey(key)
	s.waitForKey()
	return s
}

func (s *SSHConnection) sendKey(k PrivateKey) {
	sig := k.Sign(s.conn.SessionID())
	b, err := json.Marshal(sig)
	if err != nil {
		panic(err)
	}
	s.conn.SendRequest("identify", false, b)
}

func (s *SSHConnection) waitForKey() {
	for req := range s.reqs {
		if req.Type != "identify" {
			log.Printf("Received message of type %q", req.Type)
			continue
		}

		var sig Signature
		err := json.Unmarshal(req.Payload, &sig)
		if err != nil {
			log.Printf("Error unmarshalling: %v", err)
			continue

		}
		err = sig.Verify()
		if err != nil {
			log.Printf("Error verifying: %v", err)
			continue
		}
		if !bytes.Equal(sig.Signed(), s.conn.SessionID()) {
			log.Printf("wrong signed session id %x != %x", sig.Signed(), s.conn.SessionID())
			continue
		}
		s.key = sig.Key()
		break
	}
}

func (s *SSHConnection) connect() error {
	for _, k := range []string{"reachability", "receipt", "payment", "packet"} {
		err := s.connectChan(k)
		if err != nil {
			s.conn.Close()
			return err
		}
	}
	return nil
}

func (s *SSHConnection) listen() {
	go func() {
		for nc := range s.chans {
			s.lock.Lock()
			if _, ok := s.c[nc.ChannelType()]; !ok {
				nc.Reject(ssh.UnknownChannelType, "Unknown channel type")
				s.lock.Unlock()
				continue
			}
			if s.c[nc.ChannelType()] == nil {
				c, r, err := nc.Accept()
				if err != nil {
					log.Printf("Error accepting channel request: %v", err)
					s.lock.Unlock()
					continue
				}
				s.c[nc.ChannelType()] = &SSHChannel{c, r}
				s.el[nc.ChannelType()].Lock()
				s.e[nc.ChannelType()] = json.NewEncoder(c)
				s.el[nc.ChannelType()].Unlock()
				s.dl[nc.ChannelType()].Lock()
				s.d[nc.ChannelType()] = json.NewDecoder(c)
				s.dl[nc.ChannelType()].Unlock()
				s.sync <- true
			} else {
				nc.Reject(ssh.ConnectionFailed, "Connection already established")
			}
			s.lock.Unlock()
		}
	}()
	<-s.sync
	<-s.sync
	<-s.sync
	<-s.sync
	go func() {
		for range s.sync {
		}
	}()
}

func (s *SSHConnection) connectChan(name string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	c, r, err := s.conn.OpenChannel(name, nil)
	if err != nil {
		return err
	}
	s.c[name] = &SSHChannel{c, r}
	s.el[name].Lock()
	s.e[name] = json.NewEncoder(c)
	s.el[name].Unlock()
	s.dl[name].Lock()
	s.d[name] = json.NewDecoder(c)
	s.dl[name].Unlock()
	return nil
}

func (s *SSHConnection) SendMap(m BloomReachabilityMap) error {
	s.lock.Lock()
	l := s.el["reachability"]
	s.lock.Unlock()
	l.Lock()
	defer l.Unlock()
	return s.e["reachability"].Encode(m)
}

func (s *SSHConnection) ReachabilityMaps() <-chan BloomReachabilityMap {
	c := make(chan BloomReachabilityMap)
	go func() {
		s.lock.Lock()
		l := s.dl["reachability"]
		s.lock.Unlock()
		for {
			l.Lock()
			var v BloomReachabilityMap
			err := s.d["reachability"].Decode(&v)
			l.Unlock()
			if err != nil {
				if err != io.EOF {
					log.Print(err)
				}
				close(c)
				return
			} else {
				c <- v
			}
		}
	}()
	return c
}

func (s *SSHConnection) SendReceipt(r PacketReceipt) error {
	s.lock.Lock()
	l := s.el["receipt"]
	s.lock.Unlock()
	l.Lock()
	defer l.Unlock()
	return s.e["receipt"].Encode(r)
}

func (s *SSHConnection) PacketReceipts() <-chan PacketReceipt {
	c := make(chan PacketReceipt)
	go func() {
		s.lock.Lock()
		l := s.dl["receipt"]
		s.lock.Unlock()
		for {
			l.Lock()
			var v PacketReceipt
			err := s.d["receipt"].Decode(&v)
			l.Unlock()
			if err != nil {
				if err != io.EOF {
					log.Print(err)
				}
				close(c)
				return
			} else {
				c <- v
			}
		}
	}()
	return c
}

func (s *SSHConnection) SendPayment(p PaymentHash) error {
	s.lock.Lock()
	l := s.el["payment"]
	s.lock.Unlock()
	l.Lock()
	defer l.Unlock()
	return s.e["payment"].Encode(p)
}

func (s *SSHConnection) Payments() <-chan PaymentHash {
	c := make(chan PaymentHash)
	go func() {
		s.lock.Lock()
		l := s.dl["payment"]
		s.lock.Unlock()
		for {
			l.Lock()
			var v PaymentHash
			err := s.d["payment"].Decode(&v)
			l.Unlock()
			if err != nil {
				if err != io.EOF {
					log.Print(err)
				}
				close(c)
				return
			} else {
				c <- v
			}
		}
	}()
	return c
}

func (s *SSHConnection) SendPacket(p Packet) error {
	s.lock.Lock()
	l := s.el["packet"]
	s.lock.Unlock()
	l.Lock()
	defer l.Unlock()
	return s.e["packet"].Encode(p)
}

func (s *SSHConnection) Packets() <-chan Packet {
	c := make(chan Packet)
	go func() {
		s.lock.Lock()
		l := s.dl["packet"]
		s.lock.Unlock()
		for {
			l.Lock()
			var v Packet
			err := s.d["packet"].Decode(&v)
			l.Unlock()
			if err != nil {
				close(c)
				return
			} else {
				c <- v
			}
		}
	}()
	return c
}

func (s *SSHConnection) Key() PublicKey {
	return s.key
}

func (s *SSHConnection) Close() error {
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

func ListenSSH(c net.Listener, key PrivateKey) *SSHListener {
	l := &SSHListener{nil, make(chan *SSHConnection)}
	l.listen(c, key)
	return l
}

func (l *SSHListener) error(err error) {
	l.err = err
	close(l.c)
	return
}

func (l *SSHListener) listen(s net.Listener, key PrivateKey) {
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
			sc := NewSSHConnection(server, chans, reqs, key)
			sc.listen()
			l.c <- sc
		}
	}()
}

func EstablishSSH(c net.Conn, address string, key PrivateKey) (*SSHConnection, error) {
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

	client, chans, reqs, err := ssh.NewClientConn(c, address, config)
	if err != nil {
		return nil, err
	}
	sc := NewSSHConnection(client, chans, reqs, key)
	err = sc.connect()
	return sc, err
}
