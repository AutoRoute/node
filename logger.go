package node

import (
	"encoding/json"
	"io"

	"github.com/AutoRoute/node/internal"
)

type Logger struct {
	BloomFilterLog *json.Encoder
}

func NewLogger(w io.Writer) Logger {
	return Logger{json.NewEncoder(w)}
}

func (lgr Logger) LogBloomFilter(brm *internal.BloomReachabilityMap) error {
	return lgr.BloomFilterLog.Encode(brm.Conglomerate)
}
