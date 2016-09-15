package internal

import (
	"io"
)

type Logger struct {
	BloomFilterLog io.Writer
}

func NewLogger(w io.Writer) Logger {
	return Logger{w}
}

func (lgr *Logger) LogBloomFilter(brm *BloomReachabilityMap) error {
	_, err := brm.Conglomerate.WriteTo(lgr.BloomFilterLog)
	return err
}
