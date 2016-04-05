package internal

import (
	"github.com/AutoRoute/node/types"
)


// A really basic routing algorithm that just sends a packet to the first
// possible destination
type basicRouting struct {
  // Reachability for deciding where we can send it to.
  reachability *reachabilityHandler
}

func newBasicRouting(r *reachabilityHandler) *basicRouting {
  return &basicRouting {
    r,
  }
}

// Finds the next place to send a packet.
// See the routingAlgorithm interface for details.
func (b *basicRouting) FindNextHop(id types.NodeAddress,
      src types.NodeAddress) (types.NodeAddress, error) {
  possible_next, err := b.reachability.FindPossibleDests(id, src)
  if err != nil {
    return "", err
  }

  // Just naively send it to the first one.
  return possible_next[0], nil
}

// Send a packet. See routingAlgorithm interface for details.
func (b *basicRouting) SendPacket(dest DataConnection,
                                      addr types.NodeAddress,
                                      packet types.Packet) error {
  return dest.SendPacket(packet)
}
