package node

import (
	"errors"
	"log"
	"sync"
)

// Takes care of maintaining and relaying maps and insures that we know which
// interfaces can reach which addresses.
type reachabilityHandler struct {
	me         NodeAddress
	l          *sync.Mutex
	conns      map[NodeAddress]MapConnection
	maps       map[NodeAddress]ReachabilityMap
	merged_map ReachabilityMap
}

func newReachability(me NodeAddress) *reachabilityHandler {
	conns := make(map[NodeAddress]MapConnection)
	maps := make(map[NodeAddress]ReachabilityMap)
	impl := &reachabilityHandler{me, &sync.Mutex{}, conns, maps, NewBloomReachabilityMap()}
	impl.merged_map.AddEntry(me)
	return impl
}

func (m *reachabilityHandler) addMap(address NodeAddress, new_map ReachabilityMap) {
	m.l.Lock()
	defer m.l.Unlock()
	m.maps[address].Merge(new_map)
	m.merged_map.Merge(new_map)
	for addr, conn := range m.conns {
		if addr != address {
			conn.SendMap(new_map.Copy())
		}
	}
}

func (m *reachabilityHandler) AddConnection(id NodeAddress, c MapConnection) {
	// TODO(colin): This should be streamed. or something similar.
	m.l.Lock()
	defer m.l.Unlock()
	m.maps[id] = NewBloomReachabilityMap()
	m.conns[id] = c

	initial_map := m.merged_map.Copy()

	// Send all our maps
	go func() {
		m.l.Lock()
		defer m.l.Unlock()
		err := c.SendMap(initial_map)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Store all received maps
	go func() {
		for rmap := range c.ReachabilityMaps() {
			rmap.Increment()
			m.addMap(id, rmap)
		}
	}()
}

func (m *reachabilityHandler) FindNextHop(id NodeAddress) (NodeAddress, error) {
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
