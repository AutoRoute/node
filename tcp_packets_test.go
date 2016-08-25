package node

import (
	"bytes"
	"net"
	"testing"
)

func TestValidRequestPacket(t *testing.T) {
	out_req := TCPTunnelRequest{}
	var in_req TCPTunnelRequest

	in_req_wire, err := out_req.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	err = in_req.UnmarshalBinary(in_req_wire)
	if err != nil {
		t.Fatal(err)
	}

	if out_req != in_req {
		t.Fatalf("Empty structs not equal?")
	}
}

func TestInvalidRequestPacketVersion(t *testing.T) {
	out_req := TCPTunnelRequest{}
	var in_req TCPTunnelRequest

	in_req_wire, err := out_req.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	in_req_wire[0] = 1
	err = in_req.UnmarshalBinary(in_req_wire)
	if err == nil {
		t.Fatalf("Didn't catch bad version")
	}
}

func TestInvalidRequestPacketType(t *testing.T) {
	out_req := TCPTunnelRequest{}
	var in_req TCPTunnelRequest

	in_req_wire, err := out_req.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	in_req_wire[1] = 1
	err = in_req.UnmarshalBinary(in_req_wire)
	if err == nil {
		t.Fatalf("Didin't catch bad packet type")
	}
}

func TestValidResponsePacket(t *testing.T) {
	ip := net.IP([]byte{127, 0, 0, 1})
	out_resp := TCPTunnelResponse{ip}
	var in_resp TCPTunnelResponse

	in_resp_wire, err := out_resp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	err = in_resp.UnmarshalBinary(in_resp_wire)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(out_resp.ip, in_resp.ip) != 0 {
		t.Fatalf("IP addresses are not the same")
	}
}

func TestInvalidResponsePacketVersion(t *testing.T) {
	ip := net.IP([]byte{127, 0, 0, 1})
	out_resp := TCPTunnelResponse{ip}
	var in_resp TCPTunnelResponse

	in_resp_wire, err := out_resp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	in_resp_wire[0] = 1
	err = in_resp.UnmarshalBinary(in_resp_wire)
	if err == nil {
		t.Fatalf("Didn't catch bad version")
	}
}

func TestInvalidResponsePacketType(t *testing.T) {
	ip := net.IP([]byte{127, 0, 0, 1})
	out_resp := TCPTunnelResponse{ip}
	var in_resp TCPTunnelResponse

	in_resp_wire, err := out_resp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	in_resp_wire[1] = 0
	err = in_resp.UnmarshalBinary(in_resp_wire)
	if err == nil {
		t.Fatalf("Didin't catch bad packet type")
	}
}

func TestValidDataPacket(t *testing.T) {
	data := []byte{'t', 'e', 's', 't'}
	out_data := TCPTunnelData{data}
	var in_data TCPTunnelData

	in_data_wire, err := out_data.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	err = in_data.UnmarshalBinary(in_data_wire)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(out_data.data, in_data.data) != 0 {
		t.Fatalf("Data is not the same")
	}
}

func TestInvalidDataPacketVersion(t *testing.T) {
	data := []byte{'t', 'e', 's', 't'}
	out_data := TCPTunnelData{data}
	var in_data TCPTunnelData

	in_data_wire, err := out_data.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	in_data_wire[0] = 1
	err = in_data.UnmarshalBinary(in_data_wire)
	if err == nil {
		t.Fatalf("Didn't catch bad version")
	}
}

func TestInvalidDataPacketType(t *testing.T) {
	data := []byte{'t', 'e', 's', 't'}
	out_data := TCPTunnelData{data}
	var in_data TCPTunnelData

	in_data_wire, err := out_data.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	in_data_wire[1] = 0
	err = in_data.UnmarshalBinary(in_data_wire)
	if err == nil {
		t.Fatalf("Didin't catch bad packet type")
	}
}
