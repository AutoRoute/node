package node

// The brains of everything, takes in connections and wires them togethor.
type Router interface {
	AddConnection(Connection)
}

type routerImpl struct {
	// A map of public key hashes to connections
	connections map[NodeAddress]Connection
	// A map of connection id's (public key hashes) to reachability maps
	maps map[NodeAddress]ReachabilityMap
}

func (r *routerImpl) AddConnection(c Connection) {
	id := c.Key().Hash("") // give Hash() string arg
	_, duplicate := r.connections[id]
	if duplicate {
		c.Close()
		return
	}
	r.connections[id] = c

	// Curry the id since the various sub connections don't know abou it
	go r.handleReachability(id, c)
	go r.handleData(id, c)
	go r.handleReceipts(id, c)
}

func (r *routerImpl) handleReachability(id NodeAddress, c MapConnection) {
	for m := range c.ReachabilityMaps() {
		// TODO(colin): This is a horrible approximation
		r.maps[id] = m
	}
}

func (r *routerImpl) handleData(id NodeAddress, p DataConnection) {
	for packet := range p.Packets() {
		for id, c := range r.connections {
			// Ideally this call shouldn't be needed, but if we don't have
			// any maps we should still be able to route to direct neighbours
			// TODO(colin): Actually handle all the economic activity
			if c.Key().Hash("") == packet.Destination() { // give Hash() string arg
				c.SendPacket(packet)
				break
			} else if r.maps[id].IsReachable(packet.Destination()) {
				c.SendPacket(packet)
				break
			}
		}
	}
}

func (r *routerImpl) handleReceipts(id NodeAddress, p ReceiptConnection) {}
