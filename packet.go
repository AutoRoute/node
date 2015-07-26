package node

import (
	"fmt"
)

type Packet struct {
	Dest NodeAddress
	Amt  int64
	Data string
}

func (p Packet) Destination() NodeAddress {
	return p.Dest
}

func (p Packet) Hash() PacketHash {
	return PacketHash(hashstring(string(p.Dest) + string(p.Amt) + string(p.Data)))
}

func (p Packet) Amount() int64 {
	return p.Amt
}

func (p Packet) String() string {
	return fmt.Sprintf("{%x %v %v}", p.Dest, p.Amt, p.Data)
}
