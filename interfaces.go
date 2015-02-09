package node

import (
	"github.com/AutoRoute/l2"
)

// The layer two protocol takes a layer two device and returns the hash of the
// Public Key of all neighbors it can find.
type NeighborFinder interface {
	Find(l2.FrameReadWriter) <-chan string
}

type ReachabilityMap interface{}

// A layer three connection allows nodes to communicate with each other, given
// a hash and a connection. This connection sends Reachabilitymaps and proof of
// delivery.
type ControlConnection interface {
}

// The layer three control plane sends reachability information and other
// control messages.
type ControlPlane interface {
	ReadabilityMaps() <-chan ReachabilityMap
}

// The layer three data plane sends packets.
type DataPlane interface {
	SendPacket([]byte)
}
