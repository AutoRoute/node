package internal

import (
	"errors"
	"log"
	"sync"

	"github.com/AutoRoute/node/types"
)

// Takes care of maintaining and relaying maps and insures that we know which
// interfaces can reach which addresses.
type reachabilityHandler struct {
	me         types.NodeAddress
	l          *sync.Mutex
	conns      map[types.NodeAddress]MapConnection
	maps       map[types.NodeAddress]*BloomReachabilityMap
	merged_map *BloomReachabilityMap
	quit       chan bool
}

func newReachability(me types.NodeAddress) *reachabilityHandler {
	conns := make(map[types.NodeAddress]MapConnection)
	maps := make(map[types.NodeAddress]*BloomReachabilityMap)
	impl := &reachabilityHandler{
		me,
		&sync.Mutex{},
		conns,
		maps,
		NewBloomReachabilityMap(),
		make(chan bool),
	}
	impl.merged_map.AddEntry(me)
	return impl
}

func (m *reachabilityHandler) addMap(address types.NodeAddress, new_map *BloomReachabilityMap) {
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

func (m *reachabilityHandler) AddConnection(id types.NodeAddress, c MapConnection) {
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
	go m.HandleConnection(id, c)
}

func (m *reachabilityHandler) HandleConnection(id types.NodeAddress, c MapConnection) {
	for {
		select {
		case rmap, ok := <-c.ReachabilityMaps():
			if !ok {
				return
			}
			rmap.Increment()
			m.addMap(id, rmap)
		case <-m.quit:
			return
		}
	}
}

func (m *reachabilityHandler) FindNextHop(id types.NodeAddress) (types.NodeAddress, error) {
	m.l.Lock()
	defer m.l.Unlock()
	_, ok := m.conns[id]
	if ok {
		return id, nil
	}

	if id == m.me {
		return id, nil
	}

	for rid, rmap := range m.maps {
		if rmap.IsReachable(id) {
			return rid, nil
		}
	}

	return "", errors.New("Unable to find host")
}

func (m *reachabilityHandler) Close() error {
	close(m.quit)
	return nil
}
