package internal

import (
	"math"
	"testing"
	"time"

	"github.com/AutoRoute/node/types"
)

// Make sure that bandwidthEstimator works in the most basic capacity.
func TestBasic(t *testing.T) {
	// Channel that we'll send information about outgoing packets down.
	outgoing := make(chan routingDecision)
	estimator := newBandwidthEstimator(outgoing)

	// Fake node addresses.
	node1 := types.NodeAddress("A")
	node2 := types.NodeAddress("B")
	src := types.NodeAddress("src")

	// Fake packet that we'll say we sent.
	packet1 := types.Packet{node1, 1, ""}
	packet2 := types.Packet{node2, 1, ""}

	decision1 := newRoutingDecision(packet1, src, node1, 10)
	decision2 := newRoutingDecision(packet2, src, node2, 10)
	// Send a few packets in quick succession from various sources.
	for i := 0; i < 10; i++ {
		outgoing <- decision1
		time.Sleep(10 * time.Millisecond)
		outgoing <- decision2
		time.Sleep(10 * time.Millisecond)
	}

	// Since everything took the same amount of time, the weights should be equal.
	nodes := make([]types.NodeAddress, 2)
	nodes[0] = node1
	nodes[1] = node2
	weight_total := float64(0)
	for _, weight := range estimator.GetWeights(nodes) {
		if math.Abs(weight-0.5) > 0.1 {
			t.Fatalf("Expected weight of 0.5, got weight of %f\n", weight)
		}
		weight_total += weight
	}

	// The weights should also approximately add to 1.
	if math.Abs(weight_total-1.0) > 0.1 {
		t.Fatalf("Expected total weight of 1.0, got %f\n", weight_total)
	}
}

// A slightly more complicated test with varying bandwidths.
func TestDifferingBandwidths(t *testing.T) {
	// Channel that we'll send information about outgoing packets down.
	outgoing := make(chan routingDecision)
	estimator := newBandwidthEstimator(outgoing)

	// Fake node addresses.
	node1 := types.NodeAddress("A")
	node2 := types.NodeAddress("B")
	src := types.NodeAddress("src")

	// Fake packet that we'll say we sent.
	packet1 := types.Packet{node1, 1, ""}
	packet2 := types.Packet{node2, 1, ""}

	decision1 := newRoutingDecision(packet1, src, node1, 10)
	decision2 := newRoutingDecision(packet2, src, node2, 10)
	for i := 0; i < 10; i++ {
		outgoing <- decision1
		time.Sleep(10 * time.Millisecond)

		outgoing <- decision2
		time.Sleep(20 * time.Millisecond)
	}

	// The weights should be split about 2/3 and 1/3.
	nodes := make([]types.NodeAddress, 2)
	nodes[0] = node1
	nodes[1] = node2
	weight_total := float64(0)
	for _, weight := range estimator.GetWeights(nodes) {
		ok := false
		if math.Abs(weight-0.666) <= 0.1 {
			ok = true
		} else if math.Abs(weight-0.333) <= 0.1 {
			ok = true
		}
		if !ok {
			t.Fatalf("Got weight of %f\n", weight)
		}
		weight_total += weight
	}

	// The weights should also approximately add to 1.
	if math.Abs(weight_total-1.0) > 0.1 {
		t.Fatalf("Expected total weight of 1.0, got %f\n", weight_total)
	}
}

// A test to make sure it can handle weights with only partial information.
func TestPartialData(t *testing.T) {
	outgoing := make(chan routingDecision)
	estimator := newBandwidthEstimator(outgoing)

	// Try to get weights for some nodes that don't exist.
	node1 := types.NodeAddress("A")
	node2 := types.NodeAddress("B")
	node3 := types.NodeAddress("C")
	src := types.NodeAddress("src")
	nodes := make([]types.NodeAddress, 3)
	nodes[0] = node1
	nodes[1] = node2
	nodes[2] = node3

	// Fake packets that we'll say we sent.
	packet1 := types.Packet{node1, 1, ""}
	packet2 := types.Packet{node2, 1, ""}

	decision1 := newRoutingDecision(packet1, src, node1, 10)
	decision2 := newRoutingDecision(packet2, src, node2, 10)

	// We should see equally distributed weights, since we have no data on
	// anything.
	weight_total := float64(0)
	for _, weight := range estimator.GetWeights(nodes) {
		if math.Abs(weight-0.333) > 0.1 {
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
		outgoing <- decision1
		time.Sleep(10 * time.Millisecond)

		outgoing <- decision2
		time.Sleep(20 * time.Millisecond)
	}

	// The way this works is that for any node for which it lacks sufficient data,
	// it will just take 1 / number of nodes as the weight, and will calculate
	// everything else within 1 - that number.
	weight_total = 0
	for _, weight := range estimator.GetWeights(nodes) {
		ok := false
		if math.Abs(weight-0.222) <= 0.1 {
			ok = true
		} else if math.Abs(weight-0.444) <= 0.1 {
			ok = true
		} else if math.Abs(weight-0.333) <= 0.1 {
			ok = true
		}
		if !ok {
			t.Fatalf("Got weight of %f\n", weight)
		}

		weight_total += weight
	}

	// The weights should also approximately add to 1.
	if math.Abs(weight_total-1.0) > 0.1 {
		t.Fatalf("Expected total weight of 1.0, got %f\n", weight_total)
	}
}
