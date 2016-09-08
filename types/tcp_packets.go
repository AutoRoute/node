package types

import (
	"errors"
	"net"
)

// TCP tunneling request packet
// Client sends empty struct to request tunneling
type TCPTunnelRequest struct {
	Source NodeAddress
}

// Returns the TCPTunnelRequest as a byte slice.
// Adds version number and message type
//   Version number is always 0 for now
//   Message type is 0 for TCPTunnelRequest
func (req *TCPTunnelRequest) MarshalBinary() ([]byte, error) {
	return append([]byte{0, 0}, []byte(req.Source)...), nil
}

// Takes byte slice from the wire and unmarshals it.
// Checks version is 0 and message type is TCPTunnelRequest (0)
func (req *TCPTunnelRequest) UnmarshalBinary(data []byte) error {
	if data[0] != 0 {
		return errors.New("Wrong packet version")
	}

	if data[1] != 0 {
		return errors.New("Wrong packet type")
	}

	req.Source = NodeAddress(data[2:])

	return nil
}

// TCP tunneling response packet
// Server sends ip addres for client after receiving a request to tunnel
type TCPTunnelResponse struct {
	IP net.IP
}

// Returns the TCPTunnelResponse as a byte slice.
// Adds version number and message type
//   Version number is always 0 for now
//   Message type is 1 for TCPTunnelResponse
func (resp *TCPTunnelResponse) MarshalBinary() ([]byte, error) {
	return append([]byte{0, 1}, resp.IP...), nil
}

// Takes byte slice from the wire and unmarshals it.
// Checks version is 0 and message type is TCPTunnelResponse (1)
// Rest of the data in the packet is the IP address
func (resp *TCPTunnelResponse) UnmarshalBinary(data []byte) error {
	if data[0] != 0 {
		return errors.New("Wrong packet version")
	}

	if data[1] != 1 {
		return errors.New("Wrong packet type")
	}

	resp.IP = data[2:]

	return nil
}

// TCP tunneling response packet
// Data packet sent during normal transmission, after handshake
type TCPTunnelData struct {
	Data []byte
}

// Returns the TCPTunnelData as a byte slice.
// Adds version number and message type
//   Version number is always 0 for now
//   Message type is 2 for TCPTunnelData
func (d *TCPTunnelData) MarshalBinary() ([]byte, error) {
	return append([]byte{0, 2}, d.Data...), nil
}

// Takes byte slice from the wire and unmarshals it.
// Checks version is 0 and message type is TCPTunnelData (2)
// Rest of the data in the packet is the message data
func (d *TCPTunnelData) UnmarshalBinary(data []byte) error {
	if data[0] != 0 {
		return errors.New("Wrong packet version")
	}

	if data[1] != 2 {
		return errors.New("Wrong packet type")
	}

	d.Data = data[2:]

	return nil
}
