package l2

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
)

// Basic utility function to take a string and turn it into a mac address.
func MacToBytes(m string) ([]byte, error) {
	b, err := hex.DecodeString(strings.Replace(m, ":", "", -1))
	if err != nil {
		return nil, err
	}
	if len(b) != 6 {
		return nil, errors.New(fmt.Sprint("Expected mac of length 6 bytes got %d", len(b)))
	}
	return b, nil
}

func macToBytesOrDie(m string) []byte {
	b, err := MacToBytes(m)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

// A utility type to make introspecting ethernet frames easier.
type EthFrame []byte

func NewEthFrame(destination, source []byte, protocol uint16, data []byte) EthFrame {
	p := make([]byte, 12+2+len(data))
	copy(p[0:6], destination)
	copy(p[6:12], source)
	binary.BigEndian.PutUint16(p[12:14], protocol)
	copy(p[14:], data)
	return EthFrame(p)
}

func (p EthFrame) Destination() []byte {
	return p[0:6]
}

func (p EthFrame) Source() []byte {
	return p[6:12]
}

// The ethernet type. Note that if this is <1504 it is likely a length instead
// and you are communicating with an extremely non standard ethernet device.
func (p EthFrame) Type() uint16 {
	return binary.BigEndian.Uint16(p[12:14])
}

func (p EthFrame) Data() []byte {
	return p[14:]
}
