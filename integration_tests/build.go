package integration_tests

import (
	"log"
	"os"
	"os/exec"
)

func BuildBinary(path string) (string, error) {
	// Install a version with race condition checking enabled and one without.
	cmd := exec.Command("go", "install", "-race", path)
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
}

// Gets the path for the version of the autoroute binary.
func GetAutoRoutePath() string {
	path := "/tmp/autoroute"
	return path
}
