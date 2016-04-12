package internal

import (
	"log"
	"sync"
	"time"

	"github.com/AutoRoute/node/types"
)

// A bandwidth estimator estimates bandwidth between us and a neighbor and
// takes care of making intelligent routing decisions.
// NOTE: bandwidthEstimator is meant to be used as a service. It will monitor
// bandwidth in the background.
type bandwidthEstimator struct {
	// routingDecision channel that our data is coming from.
	outgoing <-chan routingDecision

	// Current estimated bandwidths.
	bandwidth      map[types.NodeAddress]float64
	bandwidth_lock sync.RWMutex

	// Total number of bytes sent.
	bytes_sent map[types.NodeAddress]int64
	// Total elapsed time waiting for that data to send.
	sent_time map[types.NodeAddress]time.Duration
	// Time at which the last packet send started.
	send_start_time map[types.NodeAddress]time.Time
	// This is so we can keep track of packets that came from the same source. It
	// maps each source to the next hop of the last packet sent from that source.
	last_sent_to map[types.NodeAddress]types.NodeAddress

	// Channel to indicate when it's time to quit.
	quit chan bool
}

// Creates a new bandwidthEstimator.
// Args:
//  routing: The routingHan.
// Returns:
//  A new bandwidthEstimator.
func newBandwidthEstimator(routing <-chan routingDecision) *bandwidthEstimator {
	estimator := bandwidthEstimator{
		routing,
		make(map[types.NodeAddress]float64),
		sync.RWMutex{},
		make(map[types.NodeAddress]int64),
		make(map[types.NodeAddress]time.Duration),
		make(map[types.NodeAddress]time.Time),
		make(map[types.NodeAddress]types.NodeAddress),
		make(chan bool),
	}

	// Start the background service.
	go estimator.monitorPackets()

	return &estimator
}

// Monitors packets that get sent through our routingHandler, and uses that data
// to glean bandwidth information.
func (b *bandwidthEstimator) monitorPackets() {
	// Channel that we use for monitoring.
	for {
		select {
		case packet, ok := <-b.outgoing:
			if !ok {
				log.Print("bandwidthEstimator: Routes channel closed, exiting.")
				return
			}
			b.sentPacket(packet.source, packet.nexthop, packet.size)
		case <-b.quit:
			return
		}
	}
}

// This should be called after a packet is sent.
// Args:
//  src: Where the packet is coming from.
//  node: Where we sent the packet.
//  size: How many bytes the packet is.
func (b *bandwidthEstimator) sentPacket(src types.NodeAddress,
	node types.NodeAddress, size int) {
	use_node := node

	last_hop, ok := b.last_sent_to[src]
	if last_hop != node {
		// If we're sending it from the same source to a different node, since each
		// source is handled by one goroutine, that means the previous send
		// completed.
		use_node = last_hop
	}

	b.last_sent_to[src] = node
	if !ok {
		// If we have no previous data, there's not much we can do in the way of
		// bandwidth.
		b.send_start_time[node] = time.Now()
		return
	}

	// Calculate new average.
	b.bytes_sent[use_node] += int64(size)

	start_time := b.send_start_time[use_node]
	elapsed := time.Since(start_time)
	b.sent_time[use_node] += elapsed

	b.bandwidth_lock.Lock()
	b.bandwidth[use_node] = float64(b.bytes_sent[use_node]) /
		(float64(b.sent_time[use_node]) / float64(time.Second))
	b.bandwidth_lock.Unlock()

	// Mark the time now. Assuming the connection is fairly saturated, this should
	// be accurate for calculating how long it took to send the next packet. If
	// it's not saturated, well, then this calculation doesn't really matter. :)
	b.send_start_time[node] = time.Now()
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
	b.bandwidth_lock.RLock()
	for _, node := range nodes {
		total += b.bandwidth[node]
	}
	b.bandwidth_lock.RUnlock()

	// For each bandwidth, scale it proportionally to get the weight.
	weights := []float64{}
	// This is the weight we use when we don't have enough data to actually
	// determine it.
	standard_weight := 1 / float64(len(nodes))
	num_unknown := 0

	// Find unknown weights first.
	for _, node := range nodes {
		b.bandwidth_lock.RLock()
		_, ok := b.bandwidth[node]
		b.bandwidth_lock.RUnlock()
		if ok {
			continue
		}

		// We don't have a valid bandwidth yet.
		weight := standard_weight
		num_unknown += 1

		b.bandwidth_lock.RLock()
		log.Printf("bandwidth: %f, weight: %f\n",
			b.bandwidth[node], weight)
		b.bandwidth_lock.RUnlock()

		weights = append(weights, weight)
	}

	// Now find the rest of them.
	for _, node := range nodes {
		b.bandwidth_lock.RLock()
		node_bandwidth, ok := b.bandwidth[node]
		b.bandwidth_lock.RUnlock()
		if !ok {
			continue
		}

		weight := float64(node_bandwidth) / total
		// If we have unknowns, scale all the weights appropriately so that everything
		// still adds up to one.
		weight *= (1 - float64(num_unknown)*standard_weight)
		b.bandwidth_lock.RLock()
		log.Printf("bandwidth: %f, weight: %f\n",
			b.bandwidth[node], weight)
		b.bandwidth_lock.RUnlock()
		weights = append(weights, weight)
	}

	return weights
}

// Stop the bandwidth estimator.
func (b *bandwidthEstimator) Close() {
	close(b.quit)
}
