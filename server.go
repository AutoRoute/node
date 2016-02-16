package node

import (
	"log"
	"net"
	"time"
)

// The server handles creating connections and listening on various ports.
type Server struct {
	n         *Node
	listeners map[string]*SSHListener
}

func NewServer(key PrivateKey, m Money) *Server {
	n := NewNode(key, m, time.Tick(30*time.Second), time.Tick(30*time.Second))
	return &Server{n, make(map[string]*SSHListener)}
}

func (s *Server) Connect(addr string) error {
	c, err := net.Dial("tcp6", addr)
	if err != nil {
		return err
	}
	m := SSHMetaData{Payment_Address: s.n.GetNewAddress()}
	sc, err := EstablishSSH(c, addr, s.n.id, m)
	if err != nil {
		return err
	}
	log.Printf("Outgoing connection: %x", sc.Key().Hash()[0:4])
	s.n.AddConnection(sc)
	return nil
}

func (s *Server) Listen(addr string) error {
	ln, err := net.Listen("tcp6", addr)
	if err != nil {
		return err
	}

	m := func() SSHMetaData {
		return SSHMetaData{Payment_Address: s.n.GetNewAddress()}
	}
	l := ListenSSH(ln, s.n.id, m)
	if l.Error() != nil {
		return err
	}
	s.listeners[addr] = l
	go func() {
		for c := range l.Connections() {
			log.Printf("Incoming connection: %x", c.Key().Hash()[0:4])
			s.n.AddConnection(c)
		}
	}()
	return nil
}

func (s *Server) AddConnection(c *SSHConnection) {
	s.n.AddConnection(c)
}

func (s *Server) Node() *Node {
	return s.n
}

func (s *Server) Close() error {
	return s.n.Close()
}
