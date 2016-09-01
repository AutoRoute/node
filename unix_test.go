package node

import (
	"bytes"
	"encoding/json"
	"net"
	"testing"

	"github.com/AutoRoute/node/internal"
	"github.com/AutoRoute/node/types"
)

func TestUnixSocket(t *testing.T) {
	_, err := NewUnixSocket("/impossible/NON EXISTANT PATH", nil)
	if err == nil {
		t.Fatal("Expected failure of opening NON EXISTANT PATH")
	}

	data := internal.TestDataConnection{make(chan types.Packet), make(chan types.Packet)}
	c, err := NewUnixSocket("/tmp/test", data)
	if err != nil {
		t.Fatal("Error opening test pipe", err)
	}
	defer c.Close()

	p := types.Packet{"dest", 10, []byte("data")}
	c2, err := net.Dial("unix", "/tmp/test")
	if err != nil {
		t.Fatal(err)
	}
	w := json.NewEncoder(c2)
	d := json.NewDecoder(c2)
	err = w.Encode(p)
	if err != nil {
		t.Fatal(err)
	}
	p2 := <-data.Out
	data.In <- p2
	var p3 types.Packet
	err = d.Decode(&p3)
	if err != nil {
		t.Fatal(err)
	}
	if p3.Dest != p.Dest || p3.Amt != p.Amt || bytes.Compare(p3.Data, p.Data) != 0 {
		t.Fatalf("Error %q != %q", p3, p)
	}
}
