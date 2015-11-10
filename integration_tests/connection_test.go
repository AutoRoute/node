package integration_tests

import (
	"testing"
	"time"
)

func TestConnection(t *testing.T) {
	listen := NewNodeBinary(BinaryOptions{listen: "localhost:9999", fake_money: true})
	listen.Start()
	defer listen.KillAndPrint(t)
	timeout := time.After(1 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		select {
		case <-timeout:
			t.Fatalf("listen binary failed to come up")
		default:
		}
		_, err := listen.GetID()
		if err == nil {
			break
		}
	}

	connect := NewNodeBinary(BinaryOptions{
		listen:     "localhost:9998",
		connect:    []string{"localhost:9999"},
		fake_money: true})
	connect.Start()
	defer connect.KillAndPrint(t)

	stop := time.After(10 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		select {
		case <-stop:
			t.Fatalf("Failed to get find connection")
		default:
		}
		connect_id, err := connect.GetID()
		if err != nil {
			continue
		}

		listen_connections, err := listen.GetConnections()
		if err != nil {
			continue
		}
		if len(listen_connections) == 0 {
			continue
		}
		if listen_connections[0] != connect_id {
			t.Fatalf("Expected %x, got %x", connect_id, listen_connections[0])
		}
		break
	}
}
