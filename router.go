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
	// TODO(colin): This should be streamed. or something similar.
	m.l.Lock()
	defer m.l.Unlock()
	m.maps[id] = nil
	m.conns[id] = c

	// Send all our maps
	go func() {
		sm := NewSimpleReachabilityMap()
		sm.AddEntry(m.me)
		for _, vs := range m.maps {
			for _, v := range vs {
				sm.Merge(v)
			}
		}
		err := c.SendMap(sm)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Store all received maps
	go func() {
		for rmap := range c.ReachabilityMaps() {
			rmap.Increment()
			m.l.Lock()
			m.maps[id] = append(m.maps[id], rmap)
			m.l.Unlock()
		}
	}()
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
	DataConnection
}

type routerImpl struct {
	// A struct which maintains reachability information
	reachability MapHandler

	id NodeAddress

	// A map of public key hashes to connections
	connections map[NodeAddress]Connection

	incoming chan Packet
}

func newRouterImpl(id NodeAddress) Router {
	return &routerImpl{newMapImpl(id), id, make(map[NodeAddress]Connection), make(chan Packet)}
}

func (r *routerImpl) AddConnection(c Connection) {
	id := c.Key().Hash()
	_, duplicate := r.connections[id]
	if duplicate {
		c.Close()
		return
	}
	r.connections[id] = c

	// Curry the id since the various sub connections don't know about it
	r.reachability.AddConnection(id, c)
	go r.handleData(id, c)
	//go r.handleReceipts(id, c)
}

func (r *routerImpl) handleData(id NodeAddress, p DataConnection) {
	for packet := range p.Packets() {
		err := r.SendPacket(packet)
		if err != nil {
			log.Printf("%q: Dropping packet destined to %q: %q", r.id, packet.Destination(), err)
			continue
		}
	}
}

func (r *routerImpl) SendPacket(p Packet) error {
	if p.Destination() == r.id {
		log.Printf("%q: Routing packet to self", r.id)
		r.incoming <- p
		return nil
	}
	rid, err := r.reachability.FindConnection(p.Destination())
	if err != nil {
		return err
	}
	log.Printf("%q: Routing packet to %q", r.id, rid)
	return r.connections[rid].SendPacket(p)
}

func (r *routerImpl) Packets() <-chan Packet {
	return r.incoming
}

//func (r *routerImpl) handleReceipts(id NodeAddress, p ReceiptConnection) {}
