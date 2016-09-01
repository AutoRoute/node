package integration_tests

import (
	"github.com/AutoRoute/node/types"

	"encoding/json"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

func WaitForID(b AutoRouteBinary) (string, error) {
	timeout := time.After(10 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		id, err := b.GetID()
		if err == nil {
			return id, nil
		}
		select {
		case <-timeout:
			return "", errors.New(fmt.Sprint("Timeout while waiting for id:", err))
		default:
		}
	}
	panic("unreachable")
}

func WaitForConnection(b AutoRouteBinary, addr string) error {
	stop := time.After(10 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		connections, err := b.GetConnections()
		if err != nil {
			continue
		}
		for _, address := range connections {
			if address == addr {
				return nil
			}
		}

		select {
		case <-stop:
			return errors.New(fmt.Sprintf("Timeout while waiting for connection to %v", addr[0:4]))
		default:
		}
	}
	panic("unreachable")
}

func WaitForSocket(dev string) (net.Conn, error) {
	timeout := time.After(time.Second)
	for range time.Tick(10 * time.Millisecond) {
		c, err := net.Dial("unix", dev)
		if err == nil {
			return c, err
		}
		select {
		case <-timeout:
			return c, err
		default:
		}
	}
	panic("Unreachable")
}

func WaitForPacketsReceived(b AutoRouteBinary, src string, amt ...int) error {
	stop := time.After(time.Second)
	for range time.Tick(10 * time.Millisecond) {
		packets_received, err := b.GetPacketsReceived()
		if err != nil {
			select {
			case <-stop:
				return err
			default:
				continue
			}
		}
		for source, amount := range packets_received {
			if len(amt) > 0 {
				if source == src && amount == amt[0] {
					return nil
				}
			} else {
				if source == src && amount > 0 {
					return nil
				}
			}
		}

		select {
		case <-stop:
			return errors.New(fmt.Sprint("Timeout out while waiting for packets received: ", packets_received))
		default:
		}
	}
	panic("unreachable")
}

func WaitForPacketsSent(b AutoRouteBinary, dest string, amt ...int) error {
	stop := time.After(time.Second)
	for range time.Tick(10 * time.Millisecond) {
		packets_sent, err := b.GetPacketsSent()
		if err != nil {
			continue
		}
		for destination, amount := range packets_sent {
			// amount should always be greater than 0
			if amount == 0 {
				return errors.New(fmt.Sprint("Invalid packet amount from destination %x", destination[0:4]))
			}
			// makes amount an optional argument
			if len(amt) > 0 {
				if destination == dest && amount == amt[0] {
					return nil
				}
			} else {
				if destination == dest {
					return nil
				}
			}
		}

		select {
		case <-stop:
			return errors.New(fmt.Sprint("Timeout out while waiting for packets sent: ", packets_sent))
		default:
		}
	}
	panic("unreachable")
}

// Waits for a single packet.
func WaitForPacket(c net.Conn, t *testing.T, s chan types.Packet) {
	WaitForPackets(c, t, s, 1)
}

// Wait for a number of packets.
// Args:
//  c: Connection to read packets from.
//  t: Testing interface being used.
//  s: Channel down which to send incomming packets.
//  amt: The amount of packets to read before exiting.
func WaitForPackets(c net.Conn, t *testing.T, s chan types.Packet, amt int) {
	r := json.NewDecoder(c)
	for i := 0; i < amt; i++ {
		var p types.Packet
		err := r.Decode(&p)
		if err != nil {
			t.Fatal(err)
		}
		s <- p
	}
}

// Sends a new packet with data we specify.
// Args:
//  conn: The connection we are sending it on.
//  t: The test interface we are using.
//  address: The address of the node where it is going.
//  data: The actual data.
func SendPacket(conn net.Conn, t *testing.T, address types.NodeAddress,
	data string) {
	// Make the packet.
	packet := types.Packet{address, 1, data}

	encoder := json.NewEncoder(conn)
	err := encoder.Encode(packet)
	if err != nil {
		t.Fatalf("Sending packet failed: %s\n", err)
	}
}
