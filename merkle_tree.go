package node

import ()

// A receipt listing packets which have been succesfully delivered
type PacketReceipt interface {
	ListPackets() []PacketHash
	Source() NodeAddress
	Verify() error
}

type Signature interface {
	Key() PublicKey
	Verify() error
	Signed() []byte
}

type merklereceipt struct {
	tree      merklenode
	signature Signature
}

type merklenode struct {
	hash  []byte
	left  *merklenode
	right *merklenode
}
