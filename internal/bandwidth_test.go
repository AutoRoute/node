package internal

import (
	"testing"
	"time"

	"github.com/AutoRoute/node/types"
)

// Make sure that bandwidthEstimator works in the most basic capacity.
func TestBasic(t *testing.T) {
	estimator := NewBandwidthEstimator()

	// Send a few packets in quick succession from various sources.
	node1 := types.NodeAddress("A")
	node2 := types.NodeAddress("B")
	for i := 0; i < 10; i++ {
		estimator.WillSendPacket(node1)
		estimator.WillSendPacket(node2)
		time.Sleep(10 * time.Millisecond)
		estimator.SentPacket(node1, 10)
		estimator.SentPacket(node2, 10)
	}

	// Since everything took the same amount of time, the weights should be equal.
	nodes := make([]types.NodeAddress, 2)
	nodes[0] = node1
	nodes[1] = node2
	weight_total := float64(0)
	for _, weight := range estimator.GetWeights(nodes) {
		if weight-0.5 > 0.001 {
			t.Fatalf("Expected weight of 0.5, got weight of %f\n", weight)
		}
		weight_total += weight
	}

	// The weights should also approximately add to 1.
	if weight_total-1.0 > 0.001 {
		t.Fatalf("Expected total weight of 1.0, got %f\n", weight_total)
	}
}

// A slightly more complicated test with varying bandwidths.
func TestDifferingBandwidths(t *testing.T) {
	estimator := NewBandwidthEstimator()

	// Send a few packets in quick succession from various sources.
	node1 := types.NodeAddress("A")
	node2 := types.NodeAddress("B")
	for i := 0; i < 10; i++ {
		estimator.WillSendPacket(node1)
		time.Sleep(10 * time.Millisecond)
		estimator.SentPacket(node1, 10)

		estimator.WillSendPacket(node2)
		time.Sleep(20 * time.Millisecond)
		estimator.SentPacket(node2, 10)
	}

	// The weights should be split about 2/3 and 1/3.
	nodes := make([]types.NodeAddress, 2)
	nodes[0] = node1
	nodes[1] = node2
	weight_total := float64(0)
	for _, weight := range estimator.GetWeights(nodes) {
		ok := false
		if weight-0.666 <= 0.001 {
			ok = true
		} else if weight-0.333 <= 0.001 {
			ok = true
		}
		if !ok {
			t.Fatalf("Got weight of %f\n", weight)
		}
		weight_total += weight
	}

	// The weights should also approximately add to 1.
	if weight_total-1.0 > 0.001 {
		t.Fatalf("Expected total weight of 1.0, got %f\n", weight_total)
	}
}

// A test to make sure it can handle weights with only partial information.
func TestPartialData(t *testing.T) {
	estimator := NewBandwidthEstimator()

	// Try to get weights for some nodes that don't exist.
	node1 := types.NodeAddress("A")
	node2 := types.NodeAddress("B")
	node3 := types.NodeAddress("C")
	nodes := make([]types.NodeAddress, 3)
	nodes[0] = node1
	nodes[1] = node2
	nodes[2] = node3

	// We should see equally distributed weights, since we have no data on
	// anything.
	weight_total := float64(0)
	for _, weight := range estimator.GetWeights(nodes) {
		if weight-0.333 > 0.001 {
			t.Fatalf("Expected weight of 0.333, got weight of %f\n", weight)
		}
		weight_total += weight
	}

	// The weights should also approximately add to 1.
	if weight_total != 1.0 {
		t.Fatalf("Expected total weight of 1.0, got %f\n", weight_total)
	}

	// Add sufficient data for two of the nodes.
	for i := 0; i < 10; i++ {
		estimator.WillSendPacket(node1)
		time.Sleep(10 * time.Millisecond)
		estimator.SentPacket(node1, 10)

		estimator.WillSendPacket(node2)
		time.Sleep(20 * time.Millisecond)
		estimator.SentPacket(node2, 10)
	}

	// The way this works is that for any node for which it lacks sufficient data,
	// it will just take 1 / number of nodes as the weight, and will calculate
	// everything else within 1 - that number.
	weight_total = 0
	for _, weight := range estimator.GetWeights(nodes) {
		ok := false
		if weight-0.222 > 0.001 {
			ok = true
		} else if weight-0.444 > 0.001 {
			ok = true
		} else if weight-0.333 > 0.001 {
			ok = true
		}
		if !ok {
			t.Fatalf("Got weight of %f\n", weight)
		}
	}

	// The weights should also approximately add to 1.
	if weight_total-1.0 > 0.001 {
		t.Fatalf("Expected total weight of 1.0, got %f\n", weight_total)
	}
}
