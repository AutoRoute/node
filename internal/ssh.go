package node

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"sync"

	"github.com/AutoRoute/node/types"

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

	// reachability
	reach_ssh_chan *SSHChannel
	reach_enc      *json.Encoder
	reach_enc_l    *sync.Mutex
	reach_dec      *json.Decoder
	reach_dec_l    *sync.Mutex
	reach_chan     chan *BloomReachabilityMap

	// receipt
	receipt_ssh_chan *SSHChannel
	receipt_enc      *json.Encoder
	receipt_enc_l    *sync.Mutex
	receipt_dec      *json.Decoder
	receipt_dec_l    *sync.Mutex
	receipt_chan     chan PacketReceipt

	// packet
	packet_ssh_chan *SSHChannel
	packet_enc      *json.Encoder
	packet_enc_l    *sync.Mutex
	packet_dec      *json.Decoder
	packet_dec_l    *sync.Mutex
	packet_chan     chan types.Packet

	other_metadata SSHMetaData
	our_metadata   SSHMetaData
	// Channel to allow blocking until ssh channels are established
	sync chan bool
	lock *sync.Mutex
}

type SSHMetaData struct {
	Payment_Address string
	// Will be generated on the fly.
	Sig Signature
}

// Constructs a new SSHConnection given the various items returned by the /x/c/ssh library.
// Does *not* call listen() or connect() on the SSHConnection, which is required
// to establish the various required channels.
func NewSSHConnection(conn ssh.Conn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request, key PrivateKey, metadata SSHMetaData) *SSHConnection {
	s := &SSHConnection{conn,
		chans,
		reqs,
		nil, nil, &sync.Mutex{}, nil, &sync.Mutex{}, make(chan *BloomReachabilityMap),
		nil, nil, &sync.Mutex{}, nil, &sync.Mutex{}, make(chan PacketReceipt),
		nil, nil, &sync.Mutex{}, nil, &sync.Mutex{}, make(chan types.Packet),
		SSHMetaData{},
		metadata,
		make(chan bool),
		&sync.Mutex{},
	}
	go s.sendMetaData(key, metadata)
	s.waitForMetaData()
	return s
}

func (s *SSHConnection) sendMetaData(key PrivateKey, m SSHMetaData) {
	m.Sig = key.Sign(s.conn.SessionID())
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	s.conn.SendRequest("identify", false, b)
}

func (s *SSHConnection) waitForMetaData() {
	for req := range s.reqs {
		if req.Type != "identify" {
			log.Printf("Received message of type %q", req.Type)
			continue
		}

		var m SSHMetaData
		err := json.Unmarshal(req.Payload, &m)
		if err != nil {
			log.Printf("Error unmarshalling: %v", err)
			continue

		}
		err = m.Sig.Verify()
		if err != nil {
			log.Printf("Error verifying: %v", err)
			continue
		}
		if !bytes.Equal(m.Sig.Signed(), s.conn.SessionID()) {
			log.Printf("wrong signed session id %x != %x", m.Sig.Signed(), s.conn.SessionID())
			continue
		}
		s.other_metadata = m
		break
	}
}

func (s *SSHConnection) listen() {
	go func() {
		for nc := range s.chans {
			s.lock.Lock()
			switch nc.ChannelType() {
			case "reachability":
				if s.reach_ssh_chan != nil {
					nc.Reject(ssh.ConnectionFailed, "Connection already established")
					s.lock.Unlock()
					continue
				}
				c, r, err := nc.Accept()
				if err != nil {
					log.Printf("Error accepting channel request: %v", err)
					s.lock.Unlock()
					continue
				}
				s.reach_ssh_chan = &SSHChannel{c, r}
				s.reach_enc_l.Lock()
				s.reach_enc = json.NewEncoder(c)
				s.reach_enc_l.Unlock()
				s.reach_dec_l.Lock()
				s.reach_dec = json.NewDecoder(c)
				s.reach_dec_l.Unlock()
				s.sync <- true
			case "receipt":
				if s.receipt_ssh_chan != nil {
					nc.Reject(ssh.ConnectionFailed, "Connection already established")
					s.lock.Unlock()
					continue
				}
				c, r, err := nc.Accept()
				if err != nil {
					log.Printf("Error accepting channel request: %v", err)
					s.lock.Unlock()
					continue
				}
				s.receipt_ssh_chan = &SSHChannel{c, r}
				s.receipt_enc_l.Lock()
				s.receipt_enc = json.NewEncoder(c)
				s.receipt_enc_l.Unlock()
				s.receipt_dec_l.Lock()
				s.receipt_dec = json.NewDecoder(c)
				s.receipt_dec_l.Unlock()
				s.sync <- true
			case "packet":
				if s.packet_ssh_chan != nil {
					nc.Reject(ssh.ConnectionFailed, "Connection already established")
					s.lock.Unlock()
					continue
				}
				c, r, err := nc.Accept()
				if err != nil {
					log.Printf("Error accepting channel request: %v", err)
					s.lock.Unlock()
					continue
				}
				s.packet_ssh_chan = &SSHChannel{c, r}
				s.packet_enc_l.Lock()
				s.packet_enc = json.NewEncoder(c)
				s.packet_enc_l.Unlock()
				s.packet_dec_l.Lock()
				s.packet_dec = json.NewDecoder(c)
				s.packet_dec_l.Unlock()
				s.sync <- true
			default:
				nc.Reject(ssh.UnknownChannelType, "Unknown channel type")
			}
			s.lock.Unlock()
		}
	}()
	<-s.sync
	<-s.sync
	<-s.sync
	go s.handleMaps()
	go s.handleReceipts()
	go s.handlePackets()
	go func() {
		for range s.sync {
		}
	}()
}

func (s *SSHConnection) connect() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	c, r, err := s.conn.OpenChannel("reachability", nil)
	if err != nil {
		return err
	}
	s.reach_ssh_chan = &SSHChannel{c, r}
	s.reach_enc_l.Lock()
	s.reach_enc = json.NewEncoder(c)
	s.reach_enc_l.Unlock()
	s.reach_dec_l.Lock()
	s.reach_dec = json.NewDecoder(c)
	s.reach_dec_l.Unlock()

	c, r, err = s.conn.OpenChannel("receipt", nil)
	if err != nil {
		return err
	}
	s.receipt_ssh_chan = &SSHChannel{c, r}
	s.receipt_enc_l.Lock()
	s.receipt_enc = json.NewEncoder(c)
	s.receipt_enc_l.Unlock()
	s.receipt_dec_l.Lock()
	s.receipt_dec = json.NewDecoder(c)
	s.receipt_dec_l.Unlock()

	c, r, err = s.conn.OpenChannel("packet", nil)
	if err != nil {
		return err
	}
	s.packet_ssh_chan = &SSHChannel{c, r}
	s.packet_enc_l.Lock()
	s.packet_enc = json.NewEncoder(c)
	s.packet_enc_l.Unlock()
	s.packet_dec_l.Lock()
	s.packet_dec = json.NewDecoder(c)
	s.packet_dec_l.Unlock()

	go s.handleMaps()
	go s.handleReceipts()
	go s.handlePackets()

	return nil
}

func (s *SSHConnection) SendMap(m *BloomReachabilityMap) error {
	s.reach_enc_l.Lock()
	defer s.reach_enc_l.Unlock()
	return s.reach_enc.Encode(m)
}

func (s *SSHConnection) handleMaps() {
	for {
		s.reach_dec_l.Lock()
		var v BloomReachabilityMap
		err := s.reach_dec.Decode(&v)
		if err != nil {
			close(s.reach_chan)
			return
		} else {
			s.reach_chan <- &v
		}
		s.reach_dec_l.Unlock()
	}
}

func (s *SSHConnection) ReachabilityMaps() <-chan *BloomReachabilityMap {
	return s.reach_chan
}

func (s *SSHConnection) SendReceipt(r PacketReceipt) error {
	s.receipt_enc_l.Lock()
	defer s.receipt_enc_l.Unlock()
	return s.receipt_enc.Encode(r)
}

func (s *SSHConnection) handleReceipts() {
	for {
		s.receipt_dec_l.Lock()
		var v PacketReceipt
		err := s.receipt_dec.Decode(&v)
		if err != nil {
			close(s.receipt_chan)
			return
		} else {
			s.receipt_chan <- v
		}
		s.receipt_dec_l.Unlock()
	}
}

func (s *SSHConnection) PacketReceipts() <-chan PacketReceipt {
	return s.receipt_chan
}

func (s *SSHConnection) SendPacket(p types.Packet) error {
	s.packet_enc_l.Lock()
	defer s.packet_enc_l.Unlock()
	return s.packet_enc.Encode(p)
}

func (s *SSHConnection) handlePackets() {
	for {
		s.packet_dec_l.Lock()
		var v types.Packet
		err := s.packet_dec.Decode(&v)
		if err != nil {
			close(s.packet_chan)
			return
		} else {
			s.packet_chan <- v
		}
		s.packet_dec_l.Unlock()
	}
}

func (s *SSHConnection) Packets() <-chan types.Packet {
	return s.packet_chan
}

func (s *SSHConnection) Key() PublicKey {
	return s.other_metadata.Sig.Key()
}

func (s *SSHConnection) MetaData() SSHMetaData {
	return s.our_metadata
}

func (s *SSHConnection) OtherMetaData() SSHMetaData {
	return s.other_metadata
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

func ListenSSH(c net.Listener, key PrivateKey, metadata func() SSHMetaData) *SSHListener {
	l := &SSHListener{nil, make(chan *SSHConnection)}
	l.listen(c, key, metadata)
	return l
}

func (l *SSHListener) error(err error) {
	l.err = err
	close(l.c)
	return
}

func (l *SSHListener) listen(s net.Listener, key PrivateKey, metadata func() SSHMetaData) {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{}, nil
		},
	}
	signer, err := ssh.NewSignerFromKey(key.K)
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
			sc := NewSSHConnection(server, chans, reqs, key, metadata())
			sc.listen()
			l.c <- sc
		}
	}()
}

func EstablishSSH(c net.Conn, address string, key PrivateKey, metadata SSHMetaData) (*SSHConnection, error) {
	username := string(key.PublicKey().Hash())
	signer, err := ssh.NewSignerFromKey(key.K)
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
	sc := NewSSHConnection(client, chans, reqs, key, metadata)
	err = sc.connect()
	if err != nil {
		sc.conn.Close()
	}
	return sc, err
}
