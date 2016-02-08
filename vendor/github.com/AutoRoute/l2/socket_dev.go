package l2

import (
	"encoding/binary"
	"io"
)

// This type is a generic wrapper around a io.ReadWriteCloser which allows
// ethernet frames to be tunneled over it.
type socketDevice struct {
	io.ReadWriter
}

// A utility function to transform a ReadWriter into a FrameReadWriter.
func WrapReadWriter(rw io.ReadWriter) FrameReadWriter {
	return &socketDevice{rw}
}

func (s *socketDevice) WriteFrame(data EthFrame) error {
	err := binary.Write(s, binary.BigEndian, int16(len(data)))
	if err != nil {
		return err
	}
	for written := 0; written < len(data); {
		n, err := s.Write(data[written:])
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}

func (s *socketDevice) ReadFrame() (EthFrame, error) {
	var size int16
	err := binary.Read(s, binary.BigEndian, &size)
	if err != nil {
		return nil, err
	}
	buffer := EthFrame(make([]byte, size))
	_, err = io.ReadFull(s, buffer)
	return buffer, err
}
