package internal

import (
	"log"
	"sync"

	"github.com/AutoRoute/node/types"
)

// Takes care of handling packet receipts, namely relaying them to other
// interested hosts and sending them to any objects which want to take action
// on them via the ReceiptAction interface.
type receiptHandler struct {
	connections map[types.NodeAddress]ReceiptConnection
	packets     map[types.PacketHash]routingDecision
	l           *sync.Mutex
	id          types.NodeAddress
	outgoing    chan types.PacketHash
	quit        chan bool
}

func newReceipt(id types.NodeAddress, c <-chan routingDecision) *receiptHandler {
	r := &receiptHandler{
		make(map[types.NodeAddress]ReceiptConnection),
		make(map[types.PacketHash]routingDecision),
		&sync.Mutex{},
		id,
		make(chan types.PacketHash),
		make(chan bool),
	}
	go r.sentPackets(c)
	return r
}

func (r *receiptHandler) AddConnection(id types.NodeAddress, c ReceiptConnection) {
	r.l.Lock()
	r.connections[id] = c
	r.l.Unlock()
	go r.handleConnection(id, c)
}

func (r *receiptHandler) handleConnection(id types.NodeAddress, c ReceiptConnection) {
	for {
		select {
		case receipt, ok := <-c.PacketReceipts():
			if !ok {
				return
			}
			r.sendReceipt(id, receipt)
		case <-r.quit:
			return
		}
	}
}

func (r *receiptHandler) sentPackets(c <-chan routingDecision) {
	for {
		select {
		case d := <-c:
			r.l.Lock()
			r.packets[d.hash] = d
			r.l.Unlock()
		case <-r.quit:
			return
		}
	}
}

func (r *receiptHandler) PacketHashes() <-chan types.PacketHash {
	return r.outgoing
}

func (r *receiptHandler) SendReceipt(receipt PacketReceipt) {
	r.sendReceipt(r.id, receipt)
}

func (r *receiptHandler) sendReceipt(id types.NodeAddress, receipt PacketReceipt) {
	if err := receipt.Verify(); err != nil {
		log.Printf("Error verifying receipt from %x: %q", id, err)
		log.Print(receipt)
		return
	}
	dest := make(map[types.NodeAddress]bool)
	r.l.Lock()
	defer r.l.Unlock()
	for _, hash := range receipt.ListPackets() {
		record, ok := r.packets[hash]
		if !ok {
			log.Printf("No record found %q", hash)
			continue
		}
		if record.destination != receipt.Source() {
			log.Printf("Invalid source %q != %q", record.source, receipt.Source())
			continue
		}
		if record.nexthop != id {
			log.Printf("Received packet receipt from wrong host? %q != %q", id, record.nexthop)
		}
		dest[record.source] = true
		r.outgoing <- hash
	}
	for addr, _ := range dest {
		if addr != r.id {
			r.connections[addr].SendReceipt(receipt)
		}
	}
}

func (r *receiptHandler) Close() error {
	close(r.quit)
	return nil
}
