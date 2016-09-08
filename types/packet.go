package types

import (
	"crypto/sha512"
	"fmt"
)

// This represents a basic packet.
type Packet struct {
	// This represents the node tht the packet should go to. Note that the NodeAddress is in fact a
	// public key hash.
	Dest NodeAddress
	// The Amt is the amount of money in satoshis that will be paid upon packet delivery.
	Amt int64
	// The data is the physical data which will be sent.
	Data []byte
}

func (p Packet) Destination() NodeAddress {
	return p.Dest
}

// The PacketHash is a useful canonical representation of the packet.
func (p Packet) Hash() PacketHash {
	o := sha512.Sum512([]byte(string(p.Dest) + string(p.Data)))
	return PacketHash(string(o[0:sha512.Size]))
}

func (p Packet) Amount() int64 {
	return p.Amt
}

func (p Packet) String() string {
	return fmt.Sprintf("{%x %v %q}", p.Dest, p.Amt, p.Data)
}
