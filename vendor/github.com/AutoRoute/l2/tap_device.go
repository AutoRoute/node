package l2

import (
	"github.com/AutoRoute/tuntap"
	"log"
	"os/exec"
)

type tapDevice struct {
	dev *tuntap.Interface
}

// A Tap Device is a new networking device that this program has created. In this case
// the normal semantics are inverted, in that frames sent to the device
// are what this interface will read and vice versa. Note that this device must be closed
// when you are done using it.
func NewTapDevice(mac, dev string) (FrameReadWriteCloser, error) {
	fd, err := tuntap.Open(dev, tuntap.DevTap)
	if err != nil {
		return nil, err
	}

	ip_path, err := exec.LookPath("ip")
	if err != nil {
		return nil, err
	}

	if mac != "" {
		cmd := exec.Command(ip_path, "link", "set", "dev", dev, "address", mac)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Print("Command output:", string(output))
			return nil, err
		}
	}

	cmd := exec.Command(ip_path, "link", "set", "dev", dev, "up")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Print("Command output:", string(output))
		return nil, err
	}

	return &tapDevice{fd}, nil
}

func (t *tapDevice) String() string {
	return "TapDevice{" + t.dev.Name() + "}"
}

func (t *tapDevice) Close() error {
	return t.dev.Close()
}

func (t *tapDevice) ReadFrame() (EthFrame, error) {
	p, err := t.dev.ReadPacket()
	if err != nil {
		return nil, err
	}
	return p.Packet, nil
}

func (t *tapDevice) WriteFrame(data EthFrame) error {
	return t.dev.WritePacket(
		&tuntap.Packet{
			Protocol: int(EthFrame(data).Type()),
			Packet:   data})
}
