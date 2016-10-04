package node

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

type routingDecision struct {
	Dest       types.NodeAddress
	Next       types.NodeAddress
	PacketSize int
	Amt        int64
}

type Logger struct {
	log_enc *json.Encoder
	lock    *sync.Mutex
}

func NewLogger(w io.Writer) Logger {
	return Logger{json.NewEncoder(w), new(sync.Mutex)}
}

func (lgr Logger) LogBloomFilter(brm *internal.BloomReachabilityMap) error {
	lgr.lock.Lock()
	defer lgr.lock.Unlock()
	return lgr.log_enc.Encode(brm.Conglomerate)
}

func (lgr Logger) LogRoutingDecision(dest types.NodeAddress, next types.NodeAddress, packet_size int, amt int64) error {
	lgr.lock.Lock()
	defer lgr.lock.Unlock()
	return lgr.log_enc.Encode(routingDecision{dest, next, packet_size, amt})
}
