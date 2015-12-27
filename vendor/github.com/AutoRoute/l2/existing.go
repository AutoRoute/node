package l2

import (
	"errors"
	"net"
	"unsafe"
)

/*
#include <sys/socket.h>
#include <linux/if_packet.h>
#include <linux/if_ether.h>
#include <linux/if_arp.h>
#include <string.h>
#include <netinet/in.h>

socklen_t LLSize() {
    return sizeof(struct sockaddr_ll);
}
*/
import "C"

// Represents an existing networking interface on the system.
type existingDevice struct {
	dev  C.int
	name string
	num  C.int
}

func (e *existingDevice) String() string {
	return "existingDevice{" + e.name + "}"
}

// Opens a connection to an existing networking interface on the system.
func ConnectExistingDevice(device string) (FrameReadWriter, error) {
	sock, err := C.socket(C.AF_PACKET, C.SOCK_RAW, C.int(C.htons(C.ETH_P_ALL)))
	if err != nil {
		return nil, err
	}

	ll_addr := C.struct_sockaddr_ll{}
	i, err := net.InterfaceByName(device)
	if err != nil {
		return nil, err
	}
	ll_addr.sll_family = C.AF_PACKET
	ll_addr.sll_ifindex = C.int(i.Index)
	ll_addr.sll_protocol = C.__be16(C.htons(C.ETH_P_ALL))
	ll_addr.sll_pkttype = C.PACKET_HOST | C.PACKET_BROADCAST
	ok, err := C.bind(sock, (*C.struct_sockaddr)(unsafe.Pointer(&ll_addr)), C.LLSize())
	if err != nil {
		return nil, err
	}
	if ok != 0 {
		return nil, errors.New("bind return !ok")
	}
	return &existingDevice{sock, device, C.int(i.Index)}, nil
}

func (e *existingDevice) ReadFrame() (EthFrame, error) {
	buffer := [1523]byte{}
	n, err := C.recvfrom(e.dev, unsafe.Pointer(&buffer[0]), C.size_t(1523), 0, nil, nil)
	if err != nil {
		return nil, err
	}
	return buffer[0:n], nil
}

func (e *existingDevice) WriteFrame(data EthFrame) error {
	socket_address := C.struct_sockaddr_ll{}
	socket_address.sll_ifindex = e.num
	socket_address.sll_halen = C.ETH_ALEN
	_, err := C.memcpy(unsafe.Pointer(&socket_address.sll_addr[0]),
		unsafe.Pointer(&data[0]), C.ETH_ALEN)
	if err != nil {
		return err
	}
	n, err := C.sendto(e.dev, unsafe.Pointer(&data[0]), C.size_t(len(data)),
		0, (*C.struct_sockaddr)(unsafe.Pointer(&socket_address)), C.LLSize())
	if err != nil {
		return err
	}
	if int(n) != len(data) {
		return errors.New("sent less data then len(data)")
	}
	return nil
}
