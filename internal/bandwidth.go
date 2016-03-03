package internal

import (
	"log"
	"time"

	"github.com/AutoRoute/node/types"
)

// A bandwidth estimator estimates bandwidth between us and a neighbor and
// takes care of making intelligent routing decisions.
type bandwidthEstimator struct {
	// Current estimated bandwidths.
	bandwidth map[types.NodeAddress]int64
	// Total number of bytes sent.
	packets_sent map[types.NodeAddress]int64
	// Time at which the last packet send started.
	send_start_time map[types.NodeAddress]time.Time
}

// Creates a new bandwidthEstimator.
func newBandwidthEstimator() *bandwidthEstimator {
	estimator := bandwidthEstimator{
		make(map[types.NodeAddress]int64),
		make(map[types.NodeAddress]int64),
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
	elapsed := time.Now().Sub(b.send_start_time[node])
	// Calculate bandwidth.
	bandwidth := size / int64(elapsed)
	b.packets_sent[node] += 1

	// Figure that into our moving average.
	total := b.bandwidth[node] * b.packets_sent[node]
	b.bandwidth[node] = total + bandwidth - b.bandwidth[node]
}

// Calculates and returns weights for each node in order to make routing
// decisions.
// Args:
//  nodes: A slice of the nodes under consideration.
// Returns:
//  A slice with the weights of each node in the same order.
func (b *bandwidthEstimator) GetWeights(nodes []types.NodeAddress) []float64 {
	// First, we need to calculate the total bandwidth so we know how to scale it.
	total := int64(0)
	for _, node := range nodes {
		total += b.bandwidth[node]
	}

	// For each bandwidth, scale it proportionally to get the weight.
	weights := []float64{}
	for _, node := range nodes {
		node_bandwidth, ok := b.bandwidth[node]
		if !ok {
			// No valid bandwidth yet.
			continue
		}
		weight := float64(node_bandwidth) / float64(total)
		log.Printf("For node %d: bandwidth: %d, weight: %f\n",
			b.bandwidth[node], weight)
		weights = append(weights, weight)
	}

	return weights
}
