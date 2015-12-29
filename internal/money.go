package node

import (
	"fmt"
)

var count int

func init() {
	count = 0
}

type FakeMoney struct {
}

func (f FakeMoney) MakePayment(amount int64, destination string) (chan bool, error) {
	c := make(chan bool)
	go func() {
		c <- true
	}()
	return c, nil
}

func (f FakeMoney) GetNewAddress() (string, chan uint64, error) {
	c := make(chan uint64)
	go func() {
		c <- 1000000
	}()
	return fmt.Sprint("fake", 0), c, nil
}
