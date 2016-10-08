package node

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

// A routingDecision contains identifying information
// for a particular packet and where it was sent
type routingDecision struct {
	Dest       types.NodeAddress
	Next       types.NodeAddress
	PacketSize int
	Amt        int64
	PacketHash types.PacketHash
}

// Logger is used for logging information about what
// the node knows about the network, where packets
// are being sent, and confirmation that the packets
// are were sent (their receipts). Ideally this is all
// the information required for an intelligent routing algorithm.
type Logger struct {
	log_enc *json.Encoder
	lock    *sync.Mutex
}

func NewLogger(w io.Writer) Logger {
	return Logger{json.NewEncoder(w), new(sync.Mutex)}
}

// LogBloomFilter logs the node's conglomerate bloom filter. Should be called
// every time a new bloom filter is received.
func (lgr Logger) LogBloomFilter(brm *internal.BloomReachabilityMap) error {
	lgr.lock.Lock()
	defer lgr.lock.Unlock()
	return lgr.log_enc.Encode(brm.Conglomerate)
}

// LogRoutingDecision logs information on a sent packet. Should be called
// every time a packet is sent.
func (lgr Logger) LogRoutingDecision(dest types.NodeAddress, next types.NodeAddress, packet_size int, amt int64, packet_hash types.PacketHash) error {
	lgr.lock.Lock()
	defer lgr.lock.Unlock()
	return lgr.log_enc.Encode(routingDecision{dest, next, packet_size, amt, packet_hash})
}

// LogPacketReceipt logs a packet hash from a received receipt. Should be
// called every time a receipt is received.
func (lgr Logger) LogPacketReceipt(packet_hash types.PacketHash) error {
	lgr.lock.Lock()
	defer lgr.lock.Unlock()
	return lgr.log_enc.Encode(packet_hash)
}
