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

func WaitForSocket(p string) (net.Conn, error) {
	timeout := time.After(time.Second)
	for range time.Tick(10 * time.Millisecond) {
		c, err := net.Dial("unix", "/tmp/unix")
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

func WaitForPacket(c net.Conn, t *testing.T, s chan node.Packet) {
	r := json.NewDecoder(c)
	var p node.Packet
	err := r.Decode(&p)
	if err != nil {
		t.Fatal(err)
	}
	s <- p
}
