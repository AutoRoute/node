package integration_tests

import (
	"log"
	"os"
	"os/exec"
)

func init() {
	cmd := exec.Command("go", "install", "github.com/AutoRoute/node/autoroute")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOBIN=/tmp")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("%s:%s:", err, out)
	}
}

func GetBinaryPath() string {
	return "/tmp/autoroute"

}
