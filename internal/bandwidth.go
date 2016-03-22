package internal

import (
	"log"
	"time"

	"github.com/AutoRoute/node/types"
)

// A bandwidth estimator estimates bandwidth between us and a neighbor and
// takes care of making intelligent routing decisions.
// NOTE: bandwidthEstimator does not inherently support concurrency. Calls to
// its methods should be protected.
type bandwidthEstimator struct {
	// Current estimated bandwidths.
	bandwidth map[types.NodeAddress]float64
	// Total number of bytes sent.
	bytes_sent map[types.NodeAddress]int64
	// Total elapsed time waiting for that data to send.
	sent_time map[types.NodeAddress]time.Duration
	// Time at which the last packet send started.
	send_start_time map[types.NodeAddress]time.Time
}

// Creates a new bandwidthEstimator.
func NewBandwidthEstimator() *bandwidthEstimator {
	estimator := bandwidthEstimator{
		make(map[types.NodeAddress]float64),
		make(map[types.NodeAddress]int64),
		make(map[types.NodeAddress]time.Duration),
		make(map[types.NodeAddress]time.Time),
	}
	return &estimator
}

// This should be called every time we're about to send a packet to someone.
// Args:
//  node: Where we're going to send the packet.
func (b *bandwidthEstimator) WillSendPacket(node types.NodeAddress) {
	b.send_start_time[node] = time.Now()
}

// This should be called after a packet is sent.
// Args:
//  node: Where we sent the packet.
//  size: How many bytes the packet is.
func (b *bandwidthEstimator) SentPacket(node types.NodeAddress, size int64) {

	// Calculate new average.
	b.bytes_sent[node] += size
	elapsed := time.Since(b.send_start_time[node])
	b.sent_time[node] += elapsed
	b.bandwidth[node] = float64(b.bytes_sent[node]) /
		(float64(b.sent_time[node]) / float64(time.Second))
}

// Calculates and returns weights for each node in order to make routing
// decisions.
// Args:
//  nodes: A slice of the nodes under consideration.
// Returns:
//  A slice with the weights of each node in the same order.
func (b *bandwidthEstimator) GetWeights(nodes []types.NodeAddress) []float64 {
	// First, we need to calculate the total bandwidth so we know how to scale it.
	total := float64(0)
	for _, node := range nodes {
		total += b.bandwidth[node]
	}

	// For each bandwidth, scale it proportionally to get the weight.
	weights := []float64{}
	// This is the weight we use when we don't have enough data to actually
	// determine it.
	standard_weight := 1 / float64(len(nodes))
	num_unknown := 0
	for _, node := range nodes {
		node_bandwidth, ok := b.bandwidth[node]
		var weight float64
		if !ok {
			// We don't have a valid bandwidth yet.
			weight = standard_weight
			num_unknown += 1
		} else {
			weight = float64(node_bandwidth) / total
			// If we have unknowns, scale all the weights appropriately so that everything
			// still adds up to one.
			weight *= (1 - float64(num_unknown)*standard_weight)
		}
		log.Printf("bandwidth: %f, weight: %f\n",
			b.bandwidth[node], weight)
		weights = append(weights, weight)
	}

	return weights
}
