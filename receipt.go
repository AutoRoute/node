package node

import (
	"log"
	"sync"
)

// Takes care of recording which packets made it to the other side
type ReceiptHandler interface {
	AddConnection(NodeAddress, ReceiptConnection)
	AddSentPacket(p Packet, src, next NodeAddress)
	SendReceipt(PacketReceipt)
}

type packetRecord struct {
	destination NodeAddress
	src         NodeAddress
	next        NodeAddress
	hash        PacketHash
}

type ReceiptAction interface {
	Receipt(PacketHash)
}

type receiptImpl struct {
	connections map[NodeAddress]ReceiptConnection
	packets     map[PacketHash]packetRecord
	l           *sync.Mutex
	a           ReceiptAction
	id          NodeAddress
}

func newReceiptImpl(id NodeAddress, a ReceiptAction) ReceiptHandler {
	return &receiptImpl{make(map[NodeAddress]ReceiptConnection), make(map[PacketHash]packetRecord), &sync.Mutex{}, a, id}
}

func (r *receiptImpl) AddConnection(id NodeAddress, c ReceiptConnection) {
	r.l.Lock()
	r.connections[id] = c
	r.l.Unlock()
	go func() {
		for receipt := range c.PacketReceipts() {
			r.sendReceipt(id, receipt)
		}
	}()
}

func (r *receiptImpl) AddSentPacket(p Packet, src, next NodeAddress) {
	r.l.Lock()
	defer r.l.Unlock()
	r.packets[p.Hash()] = packetRecord{p.Destination(), src, next, p.Hash()}
}

func (r *receiptImpl) SendReceipt(receipt PacketReceipt) {
	r.sendReceipt(r.id, receipt)
}

func (r *receiptImpl) sendReceipt(id NodeAddress, receipt PacketReceipt) {
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
		r.a.Receipt(hash)
	}
	for addr, _ := range dest {
		if addr != r.id {
			r.connections[addr].SendReceipt(receipt)
		}
	}
}
