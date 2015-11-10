package integration_tests

import (
	"log"
	"testing"
)

func TestConnection(t *testing.T) {
	log.Print("making binaries")
	listen := NewNodeBinary(BinaryOptions{listen: "localhost:9999"})
	log.Print("listen start")
	listen.Start()
	defer listen.Wait()
	defer listen.Process.Kill()
	log.Print("connect start")
	connect := NewNodeBinary(BinaryOptions{listen: "localhost:9998", connect: []string{"localhost:9999"}})
	log.Print("connect start")
	connect.Start()
	defer connect.Wait()
	defer connect.Process.Kill()
	log.Print("destroying")
}
