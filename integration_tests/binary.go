package integration_tests

import (
	"os/exec"
	"strings"
)

// Represents a binary execution of the autoroute binary.
type Binary struct {
	*exec.Cmd
}

// Represents various options which can be passed to the binary.
type BinaryOptions struct {
	listen     string
	fake_money bool
	connect    []string
}

func ProduceCommandLine(b BinaryOptions) []string {
	args := make([]string, 0)
	args = append(args, "--listen="+b.listen)
	if b.fake_money {
		args = append(args, "--fake_money")
	}
	if len(b.connect) > 0 {
		args = append(args, "--connect="+strings.Join(b.connect, ","))
	}
	return args
}

func NewNodeBinary(b BinaryOptions) Binary {
	cmd := exec.Command("autoroute", ProduceCommandLine(b)...)
	return Binary{cmd}
}
