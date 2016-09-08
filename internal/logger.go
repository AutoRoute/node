package internal

import (
	"io"
)

type Logger struct {
	Log io.Writer
}

func NewLogger(w io.Writer) Logger {
	return Logger{w}
}

func (lgr *Logger) LogBloomFilter(brm *BloomReachabilityMap) error {
	_, err := brm.Conglomerate.WriteTo(lgr.Log)
	return err
}
