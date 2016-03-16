package node

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AutoRoute/l2"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

// The server handles creating connections and listening on various ports.
type Server struct {
	n         *internal.Node
	listeners map[string]*internal.SSHListener
	// This contains the addresses of the nodes that we are currently connecting
	// to, so we don't try to connect to them twice at the same time.
	currently_connecting map[string]bool
	// Mutex to protect accesses to te currently_connecting map.
	connecting_mutex sync.RWMutex
	listen_address   string
}

func NewServer(key Key, m types.Money) *Server {
	n := internal.NewNode(key.k, m, time.Tick(30*time.Second), time.Tick(30*time.Second))
	return &Server{n, make(map[string]*internal.SSHListener),
		make(map[string]bool), sync.RWMutex{}, ""}
}

func (s *Server) Connect(addr string) error {
	// Check that we should be connecting.
	s.connecting_mutex.RLock()
	_, addr_present := s.currently_connecting[addr]
	s.connecting_mutex.RUnlock()
	if addr_present {
		return errors.New(fmt.Sprintf("Already connecting to %s.\n", addr))
	}
	s.connecting_mutex.Lock()
	s.currently_connecting[addr] = true
	s.connecting_mutex.Unlock()

	c, err := net.Dial("tcp6", addr)

	// Now that we've tried connecting, remove it as a pending connection.
	s.connecting_mutex.Lock()
	delete(s.currently_connecting, addr)
	s.connecting_mutex.Unlock()

	if err != nil {
		return err
	}
	m := internal.SSHMetaData{Payment_Address: s.n.GetNewAddress()}
	sc, err := internal.EstablishSSH(c, addr, s.n.ID(), m)
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

	m := func() internal.SSHMetaData {
		return internal.SSHMetaData{Payment_Address: s.n.GetNewAddress()}
	}
	l := internal.ListenSSH(ln, s.n.ID(), m)
	if l.Error() != nil {
		return err
	}
	s.listen_address = addr
	s.listeners[addr] = l
	go func() {
		for c := range l.Connections() {
			log.Printf("Incoming connection: %x", c.Key().Hash()[0:4])
			s.n.AddConnection(c)
		}
	}()
	return nil
}

func (s *Server) Node() Node {
	return Node{s.n}
}

func (s *Server) Close() error {
	return s.n.Close()
}

func getLinkLocalAddr(dev net.Interface) (net.IP, error) {
	dev_addrs, err := dev.Addrs()
	if err != nil {
		return nil, err
	}

	for _, dev_addr := range dev_addrs {
		addr, _, err := net.ParseCIDR(dev_addr.String())
		if err != nil {
			return nil, err
		}

		if addr.IsLinkLocalUnicast() {
			return addr, nil
		}
	}
	return nil, errors.New("Unable to find Link local address")
}

func (s *Server) findNeighbors(dev net.Interface, ll_addr net.IP, port uint16) {
	conn, err := l2.ConnectExistingDevice(dev.Name)
	if err != nil {
		log.Fatal(err)
	}
	nf := internal.NewNeighborFinder(s.n.ID().PublicKey(), ll_addr, port)
	neighbors, err := nf.Find(dev.HardwareAddr, conn)
	if err != nil {
		log.Fatal(err)
	}
	for neighbor := range neighbors {
		log.Printf("Neighbour Found %x", neighbor.FullNodeAddr)
		if neighbor.FullNodeAddr == s.n.GetAddress().Hash() {
			log.Print("Warning: Ignoring connection from self.\n")
			continue
		}
		err := s.Connect(fmt.Sprintf("[%s%%%s]:%v", neighbor.LLAddrStr, dev.Name, neighbor.Port))
		if err != nil {
			log.Printf("Error connecting: %v", err)
			continue
		}
		log.Printf("Connection established to %x", neighbor.FullNodeAddr)
	}
}

func (s *Server) Probe(dev net.Interface) error {
	if dev.Name == "lo" {
		return errors.New("Unable to prove loopback device")
	}

	log.Printf("Probing %q", dev.Name)

	ll_addr, err := getLinkLocalAddr(dev)
	if err != nil {
		return err
	}

	parsed_listen_addr := strings.Split(s.listen_address, ":")
	port, err := strconv.ParseUint(parsed_listen_addr[len(parsed_listen_addr)-1], 10, 16)
	if err != nil {
		return errors.New("Unable to figure out listening port from '" + s.listen_address + "'")
	}

	go s.findNeighbors(dev, ll_addr, uint16(port))
	return nil
}
