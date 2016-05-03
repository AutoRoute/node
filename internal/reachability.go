package internal

import (
	"errors"
	"expvar"
	"log"
	"sync"

	"github.com/AutoRoute/node/types"
)

// Export the last number of possible nexthops, and the destination we were
// trying to reach.
var next_hops *expvar.Int
var destination *expvar.String

func init() {
	next_hops = expvar.NewInt("next_hops")
	destination = expvar.NewString("destination")
}

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

	if !m.maps[address].Merge(new_map) {
		// If this returns false, then we know we have already seen this map and
		// passed it along.
		log.Print("Dropping duplicate map.")
		return
	}

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
			log.Fatalf("Sending map failed: %s\n", err)
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

// Finds the set of nodes that we could send the packet to.
// Args:
//  id: The destination node.
//  src: The source node. (So we don't send it backwards.)
// Returns:
//  All the nodes that we could possibly send the packet to.
func (m *reachabilityHandler) FindPossibleDests(id types.NodeAddress,
	src types.NodeAddress) ([]types.NodeAddress, error) {
	human_address, err := id.MarshalText()
	if err != nil {
		log.Printf("Warning: Converting to human-readable address failed: %s\n", err)
	}
	destination.Set(string(human_address))

	m.l.Lock()
	defer m.l.Unlock()
	_, ok := m.conns[id]
	if ok || id == m.me {
		next_hops.Set(1)
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
		next_hops.Set(0)
		return nil, errors.New("Unable to find host")
	}
	next_hops.Set(int64(len(dests)))
	return dests, nil
}

func (m *reachabilityHandler) Close() error {
	close(m.quit)
	return nil
}
