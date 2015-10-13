package node

import (
	"log"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcrpcclient"
	"github.com/btcsuite/btcutil"
)

func NewRPCMoney(host, user, pass string) (*RPCMoney, error) {
	config := &btcrpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := btcrpcclient.New(config, nil)
	return &RPCMoney{client, time.Minute, nil}, err
}

// RPCMoney represents a Money system which is backed by a bitcoin daemon over
// an RPC connection.
type RPCMoney struct {
	rpc       *btcrpcclient.Client
	poll_rate time.Duration
	err       error
}

// Returns a channel which will receive a nil once the payment is confirmed , or
// an error if it isn't confirmed / errors.
func (r *RPCMoney) MakePayment(amount int64, destination string) (chan bool, error) {
	address, err := btcutil.DecodeAddress(destination, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	hash, err := r.rpc.SendToAddress(address, btcutil.Amount(int64(amount)))
	if err != nil {
		return nil, err
	}
	c := make(chan bool)
	go r.waitForPaymentSuccess(hash, c)
	return c, nil
}

// Waits for a payment to suceed. If it errors it will close the channel and log the error + store it in the struct.
func (r *RPCMoney) waitForPaymentSuccess(hash *wire.ShaHash, c chan bool) {
	for _ = range time.Tick(r.poll_rate) {
		info, err := r.rpc.GetTransaction(hash)
		if err != nil {
			log.Print(err)
			r.err = err
			close(c)
			return
		}
		if info.Confirmations > 6 {
			c <- true
			return
		}
	}

}

// Returns the address + a channel down which all balance changes will be sent.
func (r *RPCMoney) GetNewAddress() (string, chan uint64, error) {
	addr, err := r.rpc.GetNewAddress("")
	c := make(chan uint64)
	go r.listenBalanceChanges(addr, c)
	return addr.EncodeAddress(), c, err
}

// Listens to an address for balance changes and sends the change down a channel.
func (r *RPCMoney) listenBalanceChanges(addr btcutil.Address, c chan uint64) {
	old_amt := int64(0)
	for _ = range time.Tick(r.poll_rate) {
		amt, err := r.rpc.GetReceivedByAddress(addr)
		if err != nil {
			log.Print(err)
			r.err = err
			return
		}
		if int64(amt) > old_amt {
			c <- uint64(int64(amt) - old_amt)
			old_amt = int64(amt)
		}
	}
}
