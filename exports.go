package node

import (
	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

// Represents a money type which is fake and purely usable for testing
func FakeMoney() types.Money {
	return node.FakeMoney{}
}

// Represents a money connection to an active bitcoin server
func NewRPCMoney(host, user, pass string) (types.Money, error) {
	return node.NewRPCMoney(host, user, pass)
}
