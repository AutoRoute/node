package node

import (
	"fmt"
	"testing"
)

func FindNeighbors(t *testing.T, mac string) {
	dev := SocketDevice{}
	nf := layer2{mac}
	msg := <= Find(dev)
	fmt.Println(msg)
}
