package node

import (
	"encoding/json"
	"io"

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
	log_encoder *json.Encoder
}

func NewLogger(w io.Writer) Logger {
	return Logger{json.NewEncoder(w)}
}

func (lgr Logger) LogBloomFilter(brm *internal.BloomReachabilityMap) error {
	return lgr.log_encoder.Encode(brm.Conglomerate)
}

func (lgr Logger) LogRoutingDecision(dest types.NodeAddress, next types.NodeAddress, packet_size int, amt int64) error {
	return lgr.log_encoder.Encode(routingDecision{dest, next, packet_size, amt})
}
