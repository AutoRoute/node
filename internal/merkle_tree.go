package node

import (
	"bytes"
	"crypto/sha512"
	"errors"
	"fmt"

	"github.com/AutoRoute/node/types"
)

func CreateMerkleReceipt(key PrivateKey, packets []types.PacketHash) PacketReceipt {
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
	return PacketReceipt{old[0], key.Sign(old[0].Hash())}
}

type PacketReceipt struct {
	Tree      merklenode
	Signature Signature
}

func (m PacketReceipt) Verify() error {
	if !bytes.Equal(m.Signature.Signed(), m.Tree.Hash()) {
		return errors.New("Signature does not match contents")
	}
	return m.Signature.Verify()
}

func (m PacketReceipt) Source() types.NodeAddress {
	return m.Signature.Key().Hash()
}

func (m PacketReceipt) ListPackets() []types.PacketHash {
	return m.Tree.ListLeafs()
}

type merklenode struct {
	LeafHash types.PacketHash
	Left     *merklenode
	Right    *merklenode
}

func (m merklenode) ListLeafs() []types.PacketHash {
	if len(m.LeafHash) != 0 {
		return []types.PacketHash{m.LeafHash}
	}
	if m.Right == nil {
		return m.Left.ListLeafs()
	}
	return append(m.Left.ListLeafs(), m.Right.ListLeafs()...)
}

func (m merklenode) Hash() []byte {
	if len(m.LeafHash) != 0 {
		s := sha512.Sum512([]byte(m.LeafHash))
		return s[0:sha512.Size]
	}
	if m.Right == nil {
		s := sha512.Sum512(append(m.Left.Hash()))
		return s[0:sha512.Size]
	}
	s := sha512.Sum512(append(m.Left.Hash(), m.Right.Hash()...))
	return s[0:sha512.Size]
}

func (m merklenode) String() string {
	if len(m.LeafHash) > 0 {
		return fmt.Sprintf("{%x}", m.LeafHash)
	}
	if m.Right == nil {
		return fmt.Sprintf("{%v nil}", *m.Left)
	}
	return fmt.Sprintf("{%v %v}", *m.Left, *m.Right)
}
