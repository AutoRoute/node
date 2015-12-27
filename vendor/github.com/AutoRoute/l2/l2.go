// Package l2 is a set of utility functions for manipulating network devices
// at the layer two networking level.
package l2

import (
	"io"
)

// Something which you can read ethernet frames from. This is distinct from
// io.Reader because you cannot slice l2 ethernet frames arbitrarily.
type FrameReader interface {
	ReadFrame() (EthFrame, error)
}

// Something which you can write ethernet frames to. This is distinct from
// io.Reader because you cannot slice l2 ethernet frames arbitrarily.
type FrameWriter interface {
	WriteFrame(EthFrame) error
}

type FrameReadWriter interface {
	FrameReader
	FrameWriter
}

type FrameReadWriteCloser interface {
	FrameReader
	FrameWriter
	io.Closer
}

// Local equivalent of io.Copy, will shove frames from a FrameReader
// into a FrameWriter
func SendFrames(source FrameReader, destination FrameWriter) error {
	for {
		p, err := source.ReadFrame()
		if err != nil {
			return err
		}
		if err = destination.WriteFrame(p); err != nil {
			return err
		}
	}
}
