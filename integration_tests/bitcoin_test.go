package integration_tests

import (
	"github.com/btcsuite/btcrpcclient"

	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func WaitForGetInfo(host, user, pass string) error {
	timeout := time.After(30 * time.Second)
	for range time.Tick(1 * time.Second) {
		config := &btcrpcclient.ConnConfig{
			Host:                 host,
			User:                 user,
			Pass:                 pass,
			HTTPPostMode:         true,
			DisableTLS:           true,
			DisableAutoReconnect: true,
		}
		client, err := btcrpcclient.New(config, nil)
		if err == nil {
			_, err = client.GetInfo()
			if err == nil {
				return nil
			}
		}
		client.Shutdown()

		select {
		case <-timeout:
			return errors.New(fmt.Sprint(err))
		default:
		}
	}
	panic("unreachable")
}

func GenerateBlocks(host, user, pass string) error {
	config := &btcrpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := btcrpcclient.New(config, nil)
	if err != nil {
		return err
	}
	// 101 blocks so the first one is usable
	_, err = client.Generate(101)
	return err
}

func TestBitcoin(t *testing.T) {

	err := os.Mkdir("/tmp/1", 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("/tmp/1")
	os.Mkdir("/tmp/2", 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("/tmp/2")

	bitcoin1 := NewWrappedBinary(
		"bitcoind", "-printoconsole", "-server=1", "-regtest=1", "-dnsseed=0", "-upnp=0",
		"-port=19000", "-rpcport=19001", "-rpcallowip=0.0.0.0/0",
		"-rpcuser=admin1", "-rpcpassword=123", "-datadir=/tmp/1")
	bitcoin1.Start()
	defer bitcoin1.KillAndPrint(t)

	bitcoin2 := NewWrappedBinary(
		"bitcoind", "-printoconsole", "-server=1", "-regtest=1", "-dnsseed=0", "-upnp=0",
		"-connect=127.0.0.1:19000",
		"-port=19010", "-rpcport=19011", "-rpcallowip=0.0.0.0/0",
		"-rpcuser=admin2", "-rpcpassword=123", "-datadir=/tmp/2")
	bitcoin2.Start()
	defer bitcoin2.KillAndPrint(t)

	err = WaitForGetInfo("127.0.0.1:19001", "admin1", "123")
	if err != nil {
		t.Fatal(err)
	}
	err = WaitForGetInfo("127.0.0.1:19011", "admin2", "123")
	if err != nil {
		t.Fatal(err)
	}

	err = GenerateBlocks("127.0.0.1:19001", "admin1", "123")
	if err != nil {
		t.Fatal(err)
	}

	listen := NewNodeBinary(BinaryOptions{
		Listen:       "[::1]:9999",
		BTCHost:      "127.0.0.1:19001",
		BTCUser:      "admin1",
		BTCPass:      "123",
		RouteLogPath: "/tmp/route1.log",
	})
	listen.Start()
	defer listen.KillAndPrint(t)
	_, err = WaitForID(listen)
	if err != nil {
		t.Fatal(err)
	}

	connect := NewNodeBinary(BinaryOptions{
		Listen:       "[::1]:9998",
		Connect:      []string{"[::1]:9999"},
		BTCHost:      "127.0.0.1:19011",
		BTCUser:      "admin2",
		BTCPass:      "123",
		RouteLogPath: "/tmp/route2.log",
	})
	connect.Start()
	defer connect.KillAndPrint(t)
	connect_id, err := WaitForID(connect)
	if err != nil {
		t.Fatal(err)
	}

	err = WaitForConnection(listen, connect_id)
	if err != nil {
		t.Fatal(err)
	}
}
