package node

import (
	"errors"
	"log"
	"sync"
)

// Takes care of maintaining maps and insures that we know which interfaces are reachable where.
type MapHandler interface {
	AddConnection(NodeAddress, MapConnection)
	FindConnection(NodeAddress) (NodeAddress, error)
}

type mapImpl struct {
	me    NodeAddress
	l     *sync.Mutex
	conns map[NodeAddress]MapConnection
	maps  map[NodeAddress][]ReachabilityMap
}

func newMapImpl(me NodeAddress) MapHandler {
	conns := make(map[NodeAddress]MapConnection)
	maps := make(map[NodeAddress][]ReachabilityMap)
	return &mapImpl{me, &sync.Mutex{}, conns, maps}
}

func (m *mapImpl) AddConnection(id NodeAddress, c MapConnection) {
	m.maps[id] = nil
	m.conns[id] = c

	// Send the wort map.
	go func() {
		sm := make(map[NodeAddress]bool)
		sm[m.me] = true
		err := c.SendMap(SimpleReachabilityMap(sm))
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Store all received maps
	for rmap := range c.ReachabilityMaps() {
		// TODO(colin): This is a horrible approximation
		m.l.Lock()
		m.maps[id] = append(m.maps[id], rmap)
		m.l.Unlock()
	}
}

func (m *mapImpl) FindConnection(id NodeAddress) (NodeAddress, error) {
	m.l.Lock()
	defer m.l.Unlock()
	_, ok := m.conns[id]
	if ok {
		return id, nil
	}

	for rid, v := range m.maps {
		for _, rmap := range v {
			if rmap.IsReachable(id) {
				return rid, nil
			}
		}
	}
	return "", errors.New("Unable to find host")
}

// The brains of everything, takes in connections and wires them togethor.
type Router interface {
	AddConnection(Connection)
}

type routerImpl struct {
	// A struct which maintains reachability information
	reachability MapHandler

	// A map of public key hashes to connections
	connections map[NodeAddress]Connection
}

func (r *routerImpl) AddConnection(c Connection) {
	id := c.Key().Hash()
	_, duplicate := r.connections[id]
	if duplicate {
		c.Close()
		return
	}
	r.connections[id] = c

	// Curry the id since the various sub connections don't know abou it
	go r.reachability.AddConnection(id, c)
	go r.handleData(id, c)
	go r.handleReceipts(id, c)
}

func (r *routerImpl) handleData(id NodeAddress, p DataConnection) {
	for packet := range p.Packets() {
		rid, err := r.reachability.FindConnection(packet.Destination())
		if err != nil {
			log.Printf("Dropping packet destined to %q no known route", packet.Destination())
			continue
		}
		r.connections[rid].SendPacket(packet)
	}
}

func (r *routerImpl) handleReceipts(id NodeAddress, p ReceiptConnection) {}
