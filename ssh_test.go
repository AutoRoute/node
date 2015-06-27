package node

import (
	"testing"
)

func TestSSH(t *testing.T) {
	sk1, _ := NewECDSAKey()
	sk2, _ := NewECDSAKey()
	l := ListenSSH("127.0.0.1:9999", sk1)
	if l.Error() != nil {
		t.Fatal(l.Error())
	}

	c1, err := EstablishSSH("127.0.0.1:9999", sk2)
	if err != nil {
		t.Fatal(err)
	}

	go c1.conn.OpenChannel("foo", []byte("foo"))

	c2 := <-l.Connections()
	nc := <-c2.chans
	if nc.ChannelType() != "foo" {
		t.Fatalf("Channel type is %q != foo", nc.ChannelType())
	}
}
