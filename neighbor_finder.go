package node

import (
  "github.com/AutoRoute/l2"
)

// The layer two protocol takes a layer two device and returns the hash of the
// Public Key of all neighbors it can find.
type NeighborFinder interface {
  Find(l2.FrameReadWriter) <-chan string
}

type l2 struct { }

func (l l2) Find(l2.FrameReadWriter frw) string {
  
}
