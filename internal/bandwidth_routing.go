package internal

import (
  "log"
	"math/rand"
  "sync"

	"github.com/AutoRoute/node/types"
)

// A more advanced routing algorithm that makes decisions based on what it knows
// about the bandwidth between nodes.
type bandwidthRouting struct {
  // Reachability handler for deciding where we can send it to.
  reachability *reachabilityHandler

  // Bandwidth estimator for all nodes.
  bandwidths *bandwidthEstimator
  // Mutex for serializing bandwidthEstimator operations.
	bandwidth_mutex sync.RWMutex
}

// Helper function that, given a set of weights and a possible next hops,
// chooses one.
// Args:
//  weights: The weights to use.
//  possible_next: The possibilities for the next hop.
// Returns:
//  The address of the next hop.
func chooseNextHop(weights []float64,
                   possible_next []types.NodeAddress) types.NodeAddress {
  choice := rand.Float64()

  // We pick the last one because if it should be the last one, floating-point
	// weirdness could cause us not to pick it.
	next := possible_next[len(possible_next)-1]
	total := 0.0
	for i, weight := range weights {
		total += weight
		log.Printf("Total: %f\n", total)
		log.Printf("Choice: %f\n", choice)
		if total >= choice {
			next = possible_next[i]
			break
		}
	}

	return next
}

func newBandwidthRouting(r *reachabilityHandler) *bandwidthRouting {
	return &bandwidthRouting{
	  r,
		NewBandwidthEstimator(),
		sync.RWMutex{},
	}
}

// Finds the next place to send a packet.
// See the routingAlgorithm interface for details.
func (b *bandwidthRouting) FindNextHop(id types.NodeAddress,
      src types.NodeAddress) (types.NodeAddress, error) {
  possible_next, err := b.reachability.FindPossibleDests(id, src)
	if err != nil {
		return "", err
	}

	// Decide which one of our possible destinations to send it to.
	b.bandwidth_mutex.RLock()
	weights := b.bandwidths.GetWeights(possible_next)
	b.bandwidth_mutex.RUnlock()

	// Choose a random destination given our weights.
	return chooseNextHop(weights, possible_next), nil
}

// Send a packet. See routingAlgorithm interface for details.
func (b *bandwidthRouting) SendPacket(dest DataConnection,
                                      addr types.NodeAddress,
                                      packet types.Packet) error {
	b.bandwidth_mutex.Lock()
	b.bandwidths.WillSendPacket(addr)
	b.bandwidth_mutex.Unlock()

	err := dest.SendPacket(packet)
	if err != nil {
		log.Print("Error sending packet.\n")
		return err
	}
	b.bandwidth_mutex.Lock()
	b.bandwidths.SentPacket(addr, int64(len(packet.Data)))
	b.bandwidth_mutex.Unlock()
	return nil
}
