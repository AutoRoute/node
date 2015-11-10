package integration_tests

import (
	"errors"
	"testing"
	"time"
)

func WaitForID(b Binary) (string, error) {
	timeout := time.After(1 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		select {
		case <-timeout:
			return "", errors.New("Timeout while waiting for id")
		default:
		}
		id, err := b.GetID()
		if err != nil {
			continue
		}
		return id, nil
	}
	panic("unreachable")
}

func WaitForConnection(b Binary, addr string) error {
	stop := time.After(10 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		select {
		case <-stop:
			return errors.New("Timeout out while waiting for connection")
		default:
		}

		connections, err := b.GetConnections()
		if err != nil {
			continue
		}
		for _, address := range connections {
			if address == addr {
				return nil
			}
		}
	}
	panic("unreachable")
}

func TestConnection(t *testing.T) {
	listen := NewNodeBinary(BinaryOptions{listen: "localhost:9999", fake_money: true})
	listen.Start()
	defer listen.KillAndPrint(t)
	_, err := WaitForID(listen)
	if err != nil {
		t.Fatal(err)
	}

	connect := NewNodeBinary(BinaryOptions{
		listen:     "localhost:9998",
		connect:    []string{"localhost:9999"},
		fake_money: true})
	connect.Start()
	defer connect.KillAndPrint(t)
	connect_id, err := WaitForID(connect)

	err = WaitForConnection(listen, connect_id)
	if err != nil {
		t.Fatal(err)
	}
}
