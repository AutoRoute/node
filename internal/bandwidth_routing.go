package internal

import (
  "fmt"
  "log"
	"math/rand"
  "sync"

	"github.com/AutoRoute/node/types"
)

// A more advanced routing handler that makes decisions based on what it knows
// about the bandwidth between nodes.
type bandwidthRouting struct {
  *basicRouting

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

func newBandwidthRouting(pk PublicKey, r *reachabilityHandler) *bandwidthRouting {
  basic_router := newBasicRouting(pk, r);

	return &bandwidthRouting{
	  basic_router,
		NewBandwidthEstimator(),
		sync.RWMutex{},
	}
}

func (r *bandwidthRouting) sendPacket(p types.Packet,
                                      src types.NodeAddress) error {
  if r.checkIfWeAreDest(p, src) {
    // We're done here.
    return nil
  }

  possible_next, err := r.reachability.FindNextHop(p.Destination(), src)
	if err != nil {
		packets_dropped.Add(1)
		return err
	}

	// Decide which one of our possible destinations to send it to.
	r.bandwidth_mutex.RLock()
	weights := r.bandwidths.GetWeights(possible_next)
	r.bandwidth_mutex.RUnlock()

	// Choose a random destination given our weights.
	next := chooseNextHop(weights, possible_next)

	packets_sent.Add(fmt.Sprintf("%x", next), 1)
	go r.notifyDecision(p, src, next)

	r.bandwidth_mutex.Lock()
	r.bandwidths.WillSendPacket(next)
	r.bandwidth_mutex.Unlock()
	err = r.connections[next].SendPacket(p)
	if err != nil {
		log.Print("Error sending packet.\n")
		return err
	}
	r.bandwidth_mutex.Lock()
	r.bandwidths.SentPacket(next, int64(len(p.Data)))
	r.bandwidth_mutex.Unlock()
	return nil
}
