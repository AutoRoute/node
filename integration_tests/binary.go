package integration_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
)

// Represents a binary execution of the autoroute binary.
type Binary struct {
	*exec.Cmd
	port   int
	output *bytes.Buffer // std err + std out
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
	port := GetUnusedPort()
	args := ProduceCommandLine(b)
	args = append(args, "--status=[::1]:"+fmt.Sprint(port))
	buf := &bytes.Buffer{}
	cmd := exec.Command(GetBinaryPath(), args...)
	cmd.Stdout = buf
	cmd.Stderr = buf
	return Binary{cmd, port, buf}
}

type statusStruct struct {
	Connections map[string]int
	Id          string
}

func (b Binary) fetchStatus() (*statusStruct, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/debug/vars", b.port))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	status := &statusStruct{}
	err = json.Unmarshal(body, status)
	return status, err
}

func (b Binary) GetConnections() ([]string, error) {
	status, err := b.fetchStatus()
	if err != nil {
		return nil, err
	}
	c := make([]string, 0)
	for k, _ := range status.Connections {
		c = append(c, k)
	}
	return c, nil

}

func (b Binary) GetID() (string, error) {
	status, err := b.fetchStatus()
	if err != nil {
		return "", err
	}
	return status.Id, nil

}

type Failer interface {
	Failed() bool
}

func (b Binary) KillAndPrint(f Failer) {
	if f.Failed() {
		fmt.Printf("\n8<----\n%s8<----\n\n", b.output)
	}
	b.Process.Kill()
}
