package integration_tests

import (
	"testing"
)

func TestConnection(t *testing.T) {
	listen := NewNodeBinary(BinaryOptions{Listen: "localhost:9999", Fake_money: true})
	listen.Start()
	defer listen.KillAndPrint(t)
	_, err := WaitForID(listen)
	if err != nil {
		t.Fatal(err)
	}

	connect := NewNodeBinary(BinaryOptions{
		Listen:     "localhost:9998",
		Connect:    []string{"localhost:9999"},
		Fake_money: true})
	connect.Start()
	defer connect.KillAndPrint(t)
	connect_id, err := WaitForID(connect)

	err = WaitForConnection(listen, connect_id)
	if err != nil {
		t.Fatal(err)
	}
}
