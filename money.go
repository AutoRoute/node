package node

import (
	"fmt"
)

var count int

func init() {
	count = 0
}

func FakeMoney() Money {
	return fakeMoney{}
}

// This represents a payment engine, which produces signed payments on demand
type Money interface {
	MakePayment(amount int64, destination string) (chan bool, error)
	GetNewAddress() (string, chan uint64, error)
}

type fakeMoney struct {
}

func (f fakeMoney) MakePayment(amount int64, destination string) (chan bool, error) {
	c := make(chan bool)
	go func() {
		c <- true
	}()
	return c, nil
}

func (f fakeMoney) GetNewAddress() (string, chan uint64, error) {
	c := make(chan uint64)
	go func() {
		c <- 1000000
	}()
	return fmt.Sprint("fake", 0), c, nil
}
