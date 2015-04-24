package node

import (
	"errors"
	"log"
	"sync"
)

// Takes care of maintaining and relaying maps and insures that we know which
// interfaces can reach which addresses.
type ReachabilityHandler interface {
	AddConnection(NodeAddress, MapConnection)
	FindNextHop(NodeAddress) (NodeAddress, error)
}

type taggedMap struct {
	address NodeAddress
	new_map ReachabilityMap
}

type reachability struct {
	me         NodeAddress
	l          *sync.Mutex
	conns      map[NodeAddress]MapConnection
	maps       map[NodeAddress]ReachabilityMap
	merged_map ReachabilityMap
}

func newReachability(me NodeAddress) ReachabilityHandler {
	conns := make(map[NodeAddress]MapConnection)
	maps := make(map[NodeAddress]ReachabilityMap)
	impl := &reachability{me, &sync.Mutex{}, conns, maps, NewBloomReachabilityMap()}
	impl.merged_map.AddEntry(me)
	return impl
}

func (m *reachability) addMap(update taggedMap) {
	m.l.Lock()
	defer m.l.Unlock()
	m.maps[update.address].Merge(update.new_map)
	m.merged_map.Merge(update.new_map)
	for addr, conn := range m.conns {
		if addr != update.address {
			conn.SendMap(update.new_map.Copy())
		}
	}
}

func (m *reachability) AddConnection(id NodeAddress, c MapConnection) {
	// TODO(colin): This should be streamed. or something similar.
	m.l.Lock()
	defer m.l.Unlock()
	m.maps[id] = NewBloomReachabilityMap()
	m.conns[id] = c

	// Send all our maps
	go func() {
		m.l.Lock()
		defer m.l.Unlock()
		err := c.SendMap(m.merged_map.Copy())
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Store all received maps
	go func() {
		for rmap := range c.ReachabilityMaps() {
			rmap.Increment()
			m.addMap(taggedMap{id, rmap})
		}
	}()
}

func (m *reachability) FindNextHop(id NodeAddress) (NodeAddress, error) {
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