package integration_tests

import (
	"errors"
	"time"
)

func WaitForID(b AutoRouteBinary) (string, error) {
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

func WaitForConnection(b AutoRouteBinary, addr string) error {
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
