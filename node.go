package node

// A Node is the highest level abstraction over the network. You receive packets
// from it and send packets to it, and it takes care of everything else.
type Node interface {
	DataConnection
}

type node struct {
}
