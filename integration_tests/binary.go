package integration_tests

import (
	"bytes"
	"os"
	"os/exec"
)

// This type represents a running binary with some nice features for test integration.
type WrappedBinary struct {
	*exec.Cmd
	output *bytes.Buffer // std err + std out
}

func NewWrappedBinary(path string, args ...string) WrappedBinary {
	buf := &bytes.Buffer{}
	cmd := exec.Command(path, args...)
	cmd.Stdout = buf
	cmd.Stderr = buf
	return WrappedBinary{cmd, buf}
}

type LogFailer interface {
	Failed() bool
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

// Used to print out the binary output in the event that some failure happened.
func (b WrappedBinary) KillAndPrint(f LogFailer) {
	if f.Failed() {
		f.Log(b.Cmd.Path)
		f.Logf("\n8<----\n%s8<----\n\n", b.output)
	}
	b.Process.Signal(os.Interrupt)
}
