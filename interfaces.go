package node

// The layer two protocol takes a layer two device and returns the hash of the
// Public Key of all neighbors it can find.
type NeighborFinder interface {
    Find(l2.FrameReadWriter) <-chan string
}

// The layer three control plane sends reachability information and other
// control messages.
type ControlPlane interface {

}
