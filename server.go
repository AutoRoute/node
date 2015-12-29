package node

import (
	"log"
	"net"
	"time"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

// The server handles creating connections and listening on various ports.
type Server struct {
	n         *fullNode
	listeners map[string]*internal.SSHListener
}

func NewServer(key Key, m types.Money) *Server {
	n := newFullNode(key.k, m, time.Tick(30*time.Second), time.Tick(30*time.Second))
	return &Server{n, make(map[string]*internal.SSHListener)}
}

func (s *Server) Connect(addr string) error {
	c, err := net.Dial("tcp6", addr)
	if err != nil {
		return err
	}
	m := internal.SSHMetaData{Payment_Address: s.n.GetNewAddress()}
	sc, err := internal.EstablishSSH(c, addr, s.n.id, m)
	if err != nil {
		return err
	}
	log.Printf("Outgoing connection: %x", sc.Key().Hash())
	s.n.AddConnection(sc)
	return nil
}

func (s *Server) Listen(addr string) error {
	ln, err := net.Listen("tcp6", addr)
	if err != nil {
		return err
	}

	m := func() internal.SSHMetaData {
		return internal.SSHMetaData{Payment_Address: s.n.GetNewAddress()}
	}
	l := internal.ListenSSH(ln, s.n.id, m)
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

func (s *Server) AddConnection(c internal.Connection) {
	s.n.AddConnection(c)
}

func (s *Server) Node() Node {
	return Node{s.n}
}

func (s *Server) Close() error {
	return s.n.Close()
}
