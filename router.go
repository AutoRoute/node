package node

// A router handles all routing tasks that don't involve the local machine
// including connection management, reachability handling, packet receipt
// relaying, and (outstanding) payment tracking. AKA anything which doesn't
// need the private key.
type Router interface {
	AddConnection(Connection)
	DataConnection
	GetAddress() PublicKey
	SendReceipt(PacketReceipt)
	SendPayment(Payment)
}

type router struct {
	pk PublicKey
	// A map of public key hashes to connections
	connections map[NodeAddress]Connection

	reachability ReachabilityHandler
	routing      RoutingHandler
	receipt      ReceiptHandler
	payment      PaymentHandler
}

func newRouter(pk PublicKey) Router {
	reach := newReachability(pk.Hash())
	routing := newRouting(pk, reach)
	c1, c2 := SplitChannel(routing.Routes())
	receipt := newReceipt(pk.Hash(), c1)
	payment := newPayment(pk.Hash(), receipt.PacketHashes(), c2)
	return &router{
		pk,
		make(map[NodeAddress]Connection),
		reach,
		routing,
		receipt,
		payment}
}

func (r *router) GetAddress() PublicKey {
	return r.pk
}

func (r *router) SendReceipt(p PacketReceipt) {
	r.receipt.SendReceipt(p)
}

func (r *router) SendPayment(p Payment) {
	r.payment.SendPayment(p)
}

func (r *router) SendPacket(p Packet) error {
	return r.routing.SendPacket(p)
}

func (r *router) Packets() <-chan Packet {
	return r.routing.Packets()
}

func (r *router) AddConnection(c Connection) {
	id := c.Key().Hash()
	_, duplicate := r.connections[id]
	if duplicate {
		c.Close()
		return
	}
	r.connections[id] = c

	// Curry the id since the various sub connections don't know about it
	r.routing.AddConnection(id, c)
	r.reachability.AddConnection(id, c)
	r.receipt.AddConnection(id, c)
	r.payment.AddConnection(id, c)
}
