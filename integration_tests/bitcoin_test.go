package integration_tests

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

var test_net_path string = filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "freewil", "bitcoin-testnet-box")

func WaitForGetInfo() error {
	timeout := time.After(10 * time.Second)
	for range time.Tick(10 * time.Millisecond) {
		cmd := exec.Command("make", "getinfo")
		cmd.Dir = test_net_path
		out, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		select {
		case <-timeout:
			return errors.New(fmt.Sprint(err, string(out)))
		default:
		}
	}
	panic("unreachable")
}

func TestBitcoin(t *testing.T) {
	log.Print("Starting")

	cmd := exec.Command("make", "start")
	cmd.Dir = test_net_path
	log.Print("Starting")
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		cmd = exec.Command("make", "stop")
		cmd.Dir = test_net_path
		log.Print("Stop")
		err = cmd.Run()
		if err != nil {
			log.Print(err)
		}
		log.Print("Done")
		cmd = exec.Command("make", "clean")
		cmd.Dir = test_net_path
		log.Print("Cleaning")
		err = cmd.Run()
		if err != nil {
			log.Print(err)
		}
		log.Print("Done")
	}()

	err = WaitForGetInfo()
	if err != nil {
		t.Fatal(err)
	}

	listen := NewNodeBinary(BinaryOptions{
		Listen:  "[::1]:9999",
		BTCHost: "[::1]:19001",
		BTCUser: "admin1",
		BTCPass: "123",
	})
	listen.Start()
	defer listen.KillAndPrint(t)
	_, err = WaitForID(listen)
	if err != nil {
		t.Fatal(err)
	}

	connect := NewNodeBinary(BinaryOptions{
		Listen:  "[::1]:9998",
		Connect: []string{"[::1]:9999"},
		BTCHost: "[::1]:19011",
		BTCUser: "admin2",
		BTCPass: "123",
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
