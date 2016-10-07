package internal

import (
	"github.com/AutoRoute/node/types"
)

type testReceiptConnection struct {
	in  chan PacketReceipt
	out chan PacketReceipt
}

func (c testReceiptConnection) SendReceipt(r PacketReceipt) error {
	c.out <- r
	return nil
}

func (c testReceiptConnection) PacketReceipts() <-chan PacketReceipt {
	return c.in
}

func makePairedReceiptConnections() (ReceiptConnection, ReceiptConnection) {
	one := make(chan PacketReceipt)
	two := make(chan PacketReceipt)
	return testReceiptConnection{one, two}, testReceiptConnection{two, one}
}

type testMapConnection struct {
	in  chan *BloomReachabilityMap
	out chan *BloomReachabilityMap
}

func (c testMapConnection) SendMap(m *BloomReachabilityMap) error {
	c.out <- m
	return nil
}

func (c testMapConnection) ReachabilityMaps() <-chan *BloomReachabilityMap {
	return c.in
}

func makePairedMapConnections() (MapConnection, MapConnection) {
	one := make(chan *BloomReachabilityMap)
	two := make(chan *BloomReachabilityMap)
	return testMapConnection{one, two}, testMapConnection{two, one}
}

type TestDataConnection struct {
	In  chan types.Packet
	Out chan types.Packet
}

func (c TestDataConnection) SendPacket(m types.Packet) error {
	c.Out <- m
	return nil
}

func (c TestDataConnection) Packets() <-chan types.Packet {
	return c.In
}

func makePairedDataConnections() (DataConnection, DataConnection) {
	one := make(chan types.Packet)
	two := make(chan types.Packet)
	return TestDataConnection{one, two}, TestDataConnection{two, one}
}

type testConnection struct {
	DataConnection
	MapConnection
	ReceiptConnection
	k PublicKey
}

func (t testConnection) Close() error               { return nil }
func (t testConnection) Key() PublicKey             { return t.k }
func (t testConnection) MetaData() SSHMetaData      { return SSHMetaData{} }
func (t testConnection) OtherMetaData() SSHMetaData { return SSHMetaData{} }

func MakePairedConnections(k1, k2 PublicKey) (Connection, Connection) {
	d1, d2 := makePairedDataConnections()
	m1, m2 := makePairedMapConnections()
	r1, r2 := makePairedReceiptConnections()
	return testConnection{d1, m1, r1, k1}, testConnection{d2, m2, r2, k2}
}

func testPacket(n types.NodeAddress) types.Packet {
	return types.Packet{n, 3, []byte("test")}
}

type Linkable interface {
	GetAddress() PublicKey
	AddConnection(Connection)
}

func Link(a, b Linkable) {
	c1, c2 := MakePairedConnections(a.GetAddress(), b.GetAddress())
	a.AddConnection(c2)
	b.AddConnection(c1)
}

type testLogger struct {
	BloomCount   int
	RouteCount   int
	ReceiptCount int
}

func (t *testLogger) LogBloomFilter(brm *BloomReachabilityMap) error {
	t.BloomCount++
	return nil
}

func (t *testLogger) LogRoutingDecision(dest types.NodeAddress, next types.NodeAddress, packet_size int, amt int64, packet_hash types.PacketHash) error {
	t.RouteCount++
	return nil
}

func (t *testLogger) LogPacketReceipt(packet_hash types.PacketHash) error {
	t.ReceiptCount++
	return nil
}
