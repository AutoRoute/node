package node

import (
	"log"
	"sync"
)

// Takes care of maintaining maps and insures that we know which interfaces are reachable where.
type ReceiptHandler interface {
	AddConnection(NodeAddress, ReceiptConnection)
	AddSentPacket(p Packet, src, next NodeAddress)
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
}

func newReceiptImpl(a ReceiptAction) ReceiptHandler {
	return &receiptImpl{make(map[NodeAddress]ReceiptConnection), make(map[PacketHash]packetRecord), &sync.Mutex{}, a}
}

func (r *receiptImpl) AddConnection(id NodeAddress, c ReceiptConnection) {
	r.l.Lock()
	r.connections[id] = c
	r.l.Unlock()
	go func() {
		for receipt := range c.PacketReceipts() {
			if receipt.Verify() != nil {
				log.Printf("Error verifying receipt: %q", receipt.Verify())
				continue
			}
			dest := make(map[NodeAddress]bool)
			r.l.Lock()
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
					log.Printf("Received packet receipt from wrong host? %q", id)
				}
				dest[record.src] = true
				r.a.Receipt(hash)
			}
			for addr, _ := range dest {
				r.connections[addr].SendReceipt(receipt)
			}
			r.l.Unlock()
		}
	}()
}

func (r *receiptImpl) AddSentPacket(p Packet, src, next NodeAddress) {
	r.l.Lock()
	defer r.l.Unlock()
	r.packets[p.Hash()] = packetRecord{p.Destination(), next, src, p.Hash()}
}
