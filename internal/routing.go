package internal

import (
	"expvar"
	"fmt"
	"log"

	"github.com/AutoRoute/node/types"
)

var packets_sent *expvar.Map
var packets_received *expvar.Map
var packets_dropped *expvar.Int

func init() {
	packets_sent = expvar.NewMap("packets_sent")
	packets_received = expvar.NewMap("packets_received")
	packets_dropped = expvar.NewInt("packets_dropped")
}

func newRoutingDecision(p types.Packet, src types.NodeAddress, nexthop types.NodeAddress) routingDecision {
	return routingDecision{p.Hash(), p.Amount(), src, p.Destination(), nexthop}
}

// A routingHandler handler takes care of relaying packets and produces notifications
// about its routingHandler decisions.
type routingHandler struct {
	pk PublicKey
	// A chan down which we send packets destined for ourselves.
	incoming chan types.Packet
	routes   chan routingDecision
	// A map of public key hashes to connections
	connections  map[types.NodeAddress]DataConnection
	quit         chan bool
	// The routingHandler algorithm to use.
	routing_algo routingAlgorithm
}

// Represents a permanent record of a routingHandler decision.
type routingDecision struct {
	hash        types.PacketHash
	amount      int64
	source      types.NodeAddress
	destination types.NodeAddress
	nexthop     types.NodeAddress
}


func newRoutingHandler(pk PublicKey,
                       algo routingAlgorithm) *routingHandler {
	return &routingHandler{
		pk,
		make(chan types.Packet),
		make(chan routingDecision),
		make(map[types.NodeAddress]DataConnection),
		make(chan bool),
		algo,
	}
}

func (r *routingHandler) AddConnection(id types.NodeAddress, c DataConnection) {
	r.connections[id] = c
	go r.handleData(id, c)
}

func (r *routingHandler) handleData(id types.NodeAddress, p DataConnection) {
	for {
		select {
		case packet, ok := <-p.Packets():
			if !ok {
				log.Printf("Packet channel closed, exiting")
				return
			}
			err := r.sendPacket(packet, id)
			if err != nil {
				log.Printf("%x: Dropping packet destined to %x: %v", r.pk.Hash(), packet.Destination(), err)
			}
		case <-r.quit:
			return
		}
	}
}

func (r *routingHandler) SendPacket(p types.Packet) error {
	return r.sendPacket(p, r.pk.Hash())
}

// Checks if it's being sent to us and handles it accordinly.
// Args:
//  p: The packet to check.
//  src: Where the packet came from.
// Returns:
//  True if it is for us. In this case, nothing else need be done. False
//  otherwise.
func (r *routingHandler) checkIfWeAreDest(p types.Packet,
                                        src types.NodeAddress) bool {
  packets_received.Add(fmt.Sprintf("%x", src), 1)
	if p.Destination() == r.pk.Hash() {
		packets_sent.Add(fmt.Sprintf("%x", r.pk.Hash()), 1)
		r.incoming <- p
		go r.notifyDecision(p, src, r.pk.Hash())
		return true
	}

	return false
}

func (r *routingHandler) sendPacket(p types.Packet, src types.NodeAddress) error {
  if r.checkIfWeAreDest(p, src) {
    // We're done here.
    return nil
  }

  next, err := r.routing_algo.FindNextHop(p.Destination(), src)
	if err != nil {
		packets_dropped.Add(1)
		return err
	}

	packets_sent.Add(fmt.Sprintf("%x", next), 1)
	go r.notifyDecision(p, src, next)

  err = r.routing_algo.SendPacket(r.connections[next], next, p)
	if err != nil {
		log.Print("Error sending packet.\n")
		return err
	}
	return nil
}

func (r *routingHandler) notifyDecision(p types.Packet, src, next types.NodeAddress) {
	r.routes <- routingDecision{p.Hash(), p.Amount(), src, p.Destination(), next}
}

func (r *routingHandler) Routes() <-chan routingDecision {
	return r.routes
}

func (r *routingHandler) Packets() <-chan types.Packet {
	return r.incoming
}

func (r *routingHandler) Close() error {
	close(r.quit)
	return nil
}
