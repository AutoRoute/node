package integration_tests

import (
	"github.com/AutoRoute/node"

	"encoding/json"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

func WaitForID(b AutoRouteBinary) (string, error) {
	timeout := time.After(2 * time.Second)
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
			return errors.New("Timeout out while waiting for connection")
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

func WaitForPacketsReceived(b AutoRouteBinary, src string, amt int) error {
	stop := time.After(time.Second)
	for range time.Tick(10 * time.Millisecond) {
		packets_received, err := b.GetPacketsReceived()
		if err != nil {
			continue
		}
		for source, amount := range packets_received {
			if source == src && amount == amt {
				return nil
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

func WaitForPacketsSent(b AutoRouteBinary, dest string, amt int) error {
	stop := time.After(time.Second)
	for range time.Tick(10 * time.Millisecond) {
		packets_sent, err := b.GetPacketsSent()
		if err != nil {
			continue
		}
		for destination, amount := range packets_sent {
			if destination == dest && amount == amt {
				return nil
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

func WaitForPacket(c net.Conn, t *testing.T, s chan node.Packet) {
	r := json.NewDecoder(c)
	var p node.Packet
	err := r.Decode(&p)
	if err != nil {
		t.Fatal(err)
	}
	s <- p
}
