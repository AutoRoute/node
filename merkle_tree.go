package node

import (
	"bytes"
	"crypto/sha512"
	"errors"
)

// A receipt listing packets which have been succesfully delivered
type PacketReceipt interface {
	ListPackets() []PacketHash
	Source() NodeAddress
	Verify() error
}

func CreateMerkleReceipt(key PrivateKey, packets []PacketHash) PacketReceipt {
	old := make([]merklenode, 0)
	for _, h := range packets {
		old = append(old, merklenode{h, nil, nil})
	}

	for {
		if len(old) <= 1 {
			break
		}
		cur := make([]merklenode, 0)
		for i, _ := range old {
			if i%2 == 1 {
				cur = append(cur, merklenode{"", &old[i], &old[i-1]})
			}
		}
		if len(old)%2 == 1 {
			cur = append(cur, merklenode{"", &old[len(old)-1], nil})
		}
		old = cur
	}
	return merklereceipt{old[0], key.Sign(old[0].Hash())}
}

type merklereceipt struct {
	tree      merklenode
	signature Signature
}

func (m merklereceipt) Verify() error {
	if !bytes.Equal(m.signature.Signed(), m.tree.Hash()) {
		return errors.New("Signature does not match contents")
	}
	return m.signature.Verify()
}

func (m merklereceipt) Source() NodeAddress {
	return m.signature.Key().Hash()
}

func (m merklereceipt) ListPackets() []PacketHash {
	return m.tree.ListLeafs()
}

type merklenode struct {
	hash  PacketHash
	left  *merklenode
	right *merklenode
}

func (m merklenode) ListLeafs() []PacketHash {
	if len(m.hash) != 0 {
		return []PacketHash{m.hash}
	}
	if m.right == nil {
		return m.left.ListLeafs()
	}
	return append(m.left.ListLeafs(), m.right.ListLeafs()...)
}

func (m merklenode) Hash() []byte {
	if len(m.hash) != 0 {
		s := sha512.Sum512([]byte(m.hash))
		return s[0:sha512.Size]
	}
	if m.right == nil {
		s := sha512.Sum512(append(m.left.Hash()))
		return s[0:sha512.Size]
	}
	s := sha512.Sum512(append(m.left.Hash(), m.right.Hash()...))
	return s[0:sha512.Size]
}
