package l2

import (
	"encoding/hex"
	"fmt"
	"log"
)

// Construct a logger which logs all frames.
func NewLogger(d FrameReader) FrameReader {
	return logger{d}
}

// Logs all frames which transit it.
type logger struct {
	D FrameReader
}

func (l logger) ReadFrame() (EthFrame, error) {
	for {
		p, err := l.D.ReadFrame()
		if err == nil {
			PrintFrame(fmt.Sprint(l.D), p)
		} else {
			log.Print("Err reading frame:", err)
		}
		return p, err
	}
}

func (l logger) String() string {
	return "Logger{" + fmt.Sprint(l.D) + "}"
}

// Utility function to pretty print a ethernet frame plus a header
func PrintFrame(name string, data []byte) {
	E := EthFrame(data)
	log.Printf("%s: %s->%s %d bytes protocol %d",
		name,
		hex.EncodeToString(E.Source()),
		hex.EncodeToString(E.Destination()),
		len(data),
		E.Type())
}
