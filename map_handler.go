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

type taggedMap struct {
	address NodeAddress
	new_map ReachabilityMap
}

type mapImpl struct {
	me         NodeAddress
	l          *sync.Mutex
	conns      map[NodeAddress]MapConnection
	maps       map[NodeAddress]ReachabilityMap
	updates    chan taggedMap
	merged_map ReachabilityMap
}

func newMapImpl(me NodeAddress) MapHandler {
	conns := make(map[NodeAddress]MapConnection)
	maps := make(map[NodeAddress]ReachabilityMap)
	impl := &mapImpl{me, &sync.Mutex{}, conns, maps, make(chan taggedMap), NewSimpleReachabilityMap()}
	go impl.maphandler()
	return impl
}

func (m *mapImpl) maphandler() {
	for update := range m.updates {
		m.l.Lock()
		m.merged_map.Merge(update.new_map)
		for addr, conn := range m.conns {
			if addr != update.address {
				conn.SendMap(update.new_map)
			}
		}
		m.l.Unlock()
	}
}

func (m *mapImpl) AddConnection(id NodeAddress, c MapConnection) {
	// TODO(colin): This should be streamed. or something similar.
	m.l.Lock()
	defer m.l.Unlock()
	m.maps[id] = NewSimpleReachabilityMap()
	m.conns[id] = c

	// Send all our maps
	go func() {
		m.l.Lock()
		defer m.l.Unlock()
		err := c.SendMap(m.merged_map)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Store all received maps
	go func() {
		for rmap := range c.ReachabilityMaps() {
			rmap.Increment()
			m.l.Lock()
			m.maps[id].Merge(rmap)
			m.updates <- taggedMap{id, rmap}
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

	for rid, rmap := range m.maps {
		if rmap.IsReachable(id) {
			return rid, nil
		}
	}
	return "", errors.New("Unable to find host")
}
