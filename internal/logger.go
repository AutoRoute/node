package internal

import (
	"encoding/json"
	"io"
)

type Logger struct {
	BloomFilterLog *json.Encoder
}

func NewLogger(w io.Writer) Logger {
	return Logger{json.NewEncoder(w)}
}

func (lgr *Logger) LogBloomFilter(brm *BloomReachabilityMap) error {
	return lgr.BloomFilterLog.Encode(brm.Conglomerate)
}
