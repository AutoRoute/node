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
	logger     Logger
	quit       chan bool
}

func newReachability(me types.NodeAddress, route_logger Logger) *reachabilityHandler {
	conns := make(map[types.NodeAddress]MapConnection)
	maps := make(map[types.NodeAddress]*BloomReachabilityMap)
	impl := &reachabilityHandler{
		me,
		&sync.Mutex{},
		conns,
		maps,
		NewBloomReachabilityMap(),
		route_logger,
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
			log.Print(err)
		}
	}()

	err := m.logger.LogBloomFilter(m.merged_map)
	if err != nil {
		log.Print(err)
	}

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

// Finds the set of nodes that we could send the packet to.
// Args:
//  id: The destination node.
//  src: The source node. (So we don't send it backwards.)
// Returns:
//  All the nodes that we could possibly send the packet to.
func (m *reachabilityHandler) FindPossibleDests(id types.NodeAddress,
	src types.NodeAddress) ([]types.NodeAddress, error) {
	m.l.Lock()
	defer m.l.Unlock()
	_, ok := m.conns[id]
	if ok {
		return []types.NodeAddress{id}, nil
	}

	if id == m.me {
		return []types.NodeAddress{id}, nil
	}

	dests := []types.NodeAddress{}
	for rid, rmap := range m.maps {
		if rid == src {
			// We're not going to send it backwards.
			continue
		}

		if rmap.IsReachable(id) {
			dests = append(dests, rid)
		}
	}

	if len(dests) == 0 {
		return nil, errors.New("Unable to find host")
	}
	return dests, nil
}

func (m *reachabilityHandler) Close() error {
	close(m.quit)
	return nil
}
