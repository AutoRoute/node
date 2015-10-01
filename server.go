package node

import (
	"fmt"
	"log"
	"net"
	"time"
)

// The server handles creating connections and listening on various ports.
type Server struct {
	n         *Node
	listeners map[string]*SSHListener
}

func NewServer(key PrivateKey) *Server {
	n := NewNode(key, time.Tick(30*time.Second), time.Tick(30*time.Second))
	return &Server{n, make(map[string]*SSHListener)}
}

func (s *Server) Connect(addr string) error {
	formatted_addr := fmt.Sprintf("[%v]:1337", addr)
	c, err := net.Dial("tcp6", formatted_addr)
	if err != nil {
		return err
	}
	sc, err := EstablishSSH(c, addr, s.n.id)
	if err != nil {
		return err
	}
	log.Printf("Outgoing connection: %x", sc.Key().Hash())
	s.n.AddConnection(sc)
	return nil
}

func (s *Server) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	l := ListenSSH(ln, s.n.id)
	if l.Error() != nil {
		return err
	}
	s.listeners[addr] = l
	go func() {
		for c := range l.Connections() {
			log.Printf("Incoming connection: %x", c.Key().Hash())
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
