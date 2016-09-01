package integration_tests

import (
	"bytes"
	"os"
	"os/exec"
	"sync"
	"time"
)

// This type represents a running binary with some nice features for test integration.
type WrappedBinary struct {
	*exec.Cmd
	buf *lockedBuffer
}

type lockedBuffer struct {
	output *bytes.Buffer // std err + std out
	lock   *sync.Mutex
}

func (l *lockedBuffer) Write(p []byte) (n int, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.output.Write(p)
}

func (l *lockedBuffer) Output() string {
	l.lock.Lock()
	defer l.lock.Unlock()
	return string(l.output.Bytes())
}

func NewWrappedBinary(path string, args ...string) WrappedBinary {
	buf := &lockedBuffer{&bytes.Buffer{}, &sync.Mutex{}}
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
		f.Logf("\n8<----\n%s8<----\n\n", b.buf.Output())
	}
	done := make(chan bool)
	go func() {
		b.Cmd.Wait()
		done <- true
	}()
	b.Process.Signal(os.Interrupt)
	select {
	case <-done:
		return
	case <-time.After(time.Second * 10):
		b.Process.Signal(os.Kill)
	}
	<-done
}
