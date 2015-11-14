package integration_tests

import (
	"log"
	"os"
	"os/exec"
)

func BuildBinary(path string) (string, error) {
	cmd := exec.Command("go", "install", path)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOBIN=/tmp")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func init() {
	out, err := BuildBinary("github.com/AutoRoute/node/autoroute")
	if err != nil {
		log.Fatalf("%s:%s:", err, out)
	}
	out, err = BuildBinary("github.com/AutoRoute/loopback2")
	if err != nil {
		log.Fatalf("%s:%s:", err, out)
	}
}

func GetAutoRoutePath() string {
	return "/tmp/autoroute"
}

func GetLoopBack2Path() string {
	return "/tmp/loopback2"
}
