package node

import (
	"log"
	"sync"
)

// Takes care of handling packet receipts, namely relaying them to other
// interested hosts and sending them to any objects which want to take action
// on them via the ReceiptAction interface.
type ReceiptHandler interface {
	AddConnection(NodeAddress, ReceiptConnection)
	SendReceipt(PacketReceipt)
	PacketHashes() <-chan PacketHash
}

type packetRecord struct {
	destination NodeAddress
	src         NodeAddress
	next        NodeAddress
	hash        PacketHash
}

type receipt struct {
	connections map[NodeAddress]ReceiptConnection
	packets     map[PacketHash]packetRecord
	l           *sync.Mutex
	id          NodeAddress
	outgoing    chan PacketHash
}

func newReceipt(id NodeAddress, c <-chan RoutingDecision) ReceiptHandler {
	r := &receipt{make(map[NodeAddress]ReceiptConnection), make(map[PacketHash]packetRecord), &sync.Mutex{}, id, make(chan PacketHash)}
	go r.sentPackets(c)
	return r
}

func (r *receipt) AddConnection(id NodeAddress, c ReceiptConnection) {
	r.l.Lock()
	r.connections[id] = c
	r.l.Unlock()
	go func() {
		for receipt := range c.PacketReceipts() {
			r.sendReceipt(id, receipt)
		}
	}()
}

func (r *receipt) sentPackets(c <-chan RoutingDecision) {
	for d := range c {
		r.l.Lock()
		r.packets[d.p.Hash()] = packetRecord{d.p.Destination(), d.source, d.nexthop, d.p.Hash()}
		r.l.Unlock()
	}
}

func (r *receipt) PacketHashes() <-chan PacketHash {
	return r.outgoing
}

func (r *receipt) SendReceipt(receipt PacketReceipt) {
	r.sendReceipt(r.id, receipt)
}

func (r *receipt) sendReceipt(id NodeAddress, receipt PacketReceipt) {
	if receipt.Verify() != nil {
		log.Printf("Error verifying receipt: %q", receipt.Verify())
		return
	}
	dest := make(map[NodeAddress]bool)
	r.l.Lock()
	defer r.l.Unlock()
	for _, hash := range receipt.ListPackets() {
		record, ok := r.packets[hash]
		if !ok {
			log.Printf("No record found %q", hash)
			continue
		}
		if record.destination != receipt.Source() {
			log.Printf("Invalid source %q != %q", record.src, receipt.Source())
			continue
		}
		if record.next != id {
			log.Printf("Received packet receipt from wrong host? %q != %q", id, record.next)
		}
		dest[record.src] = true
		log.Printf("%q: Notifying action", r.id)
		r.outgoing <- hash
	}
	for addr, _ := range dest {
		if addr != r.id {
			r.connections[addr].SendReceipt(receipt)
		}
	}
}
