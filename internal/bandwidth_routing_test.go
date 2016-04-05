package internal

import (
  "math/rand"
  "testing"

	"github.com/AutoRoute/node/types"
)

// Checks that we can properly choose the next place to send a packet.
func TestNextHopChoosing(t *testing.T) {
  // Make sure our "random" choices are deterministic for testing.
  rand.Seed(42)

  // Fake lists of hosts and weights to use for testing.
  nodes := []types.NodeAddress{"A", "B", "C", "D"}
  weights := []float64{0.2, 0.3, 0.4, 0.1}

  // We know what the output from rand will be, so make sure it chooses the
  // right ones.
  choice := chooseNextHop(weights, nodes)
  if choice != nodes[1] {
    t.Errorf("Expected node %s, got node %s.", nodes[2], choice)
  }

  choice = chooseNextHop(weights, nodes)
  if choice != nodes[0] {
    t.Errorf("Expected node %s, got node %s.", nodes[0], choice)
  }

  choice = chooseNextHop(weights, nodes)
  if choice != nodes[2] {
    t.Errorf("Expected node %s, got node %s.", nodes[2], choice)
  }
}
