package node

import (
	"fmt"
	"strconv"
	"strings"
)

// A type representing information about a valid payment
type Payment struct {
	Source      NodeAddress
	Destination NodeAddress
	Amount      int64
}

// A type representing a hash of a payment which you are telling the other side to use.
type PaymentHash string

// This represents a payment engine, which produces signed payments on demand
type Money interface {
	MakePayment(amount int64, destination NodeAddress) (PaymentHash, error)
	AddPaymentHash(PaymentHash) chan Payment
}

type fakeMoney struct {
	id NodeAddress
}

func (f fakeMoney) MakePayment(amount int64, destination NodeAddress) (PaymentHash, error) {
	return PaymentHash(fmt.Sprintf("%s:%d", destination, amount)), nil
}

func (f fakeMoney) AddPaymentHash(p PaymentHash) chan Payment {
	c := make(chan Payment)
	v := strings.Split(string(p), ":")
	a, _ := strconv.Atoi(v[1])
	go func() {
		c <- Payment{NodeAddress(v[0]), f.id, int64(a)}
	}()
	return c
}
