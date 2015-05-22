package node

// A type representing a payment that you can use
type Payment interface {
	Source() NodeAddress
	Destination() NodeAddress
	Verify() error
	Amount() int64
}

type testPayment struct {
	src NodeAddress
	dst NodeAddress
	amt int64
}

func (t testPayment) Source() NodeAddress      { return t.src }
func (t testPayment) Destination() NodeAddress { return t.dst }
func (t testPayment) Verify() error            { return nil }
func (t testPayment) Amount() int64            { return t.amt }

// This represents a payment engine, which produces signed payments on demand
type Money interface {
	MakePayment(amount int64, destination NodeAddress) (Payment, error)
	HandlePayment(Payment) chan error
}

type fakeMoney struct {
	id NodeAddress
}

func (f fakeMoney) MakePayment(amount int64, destination NodeAddress) (Payment, error) {
	return testPayment{f.id, destination, amount}, nil
}

func (f fakeMoney) HandlePayment(p Payment) chan error {
	c := make(chan error)
	go func() {
		c <- nil
	}()
	return c
}
