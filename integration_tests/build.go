package integration_tests

import (
	"log"
	"os"
	"os/exec"
)

func BuildBinary(path string) (string, error) {
	// Install a version with race condition checking enabled and one without.
	cmd := exec.Command("go", "install", path)
	race_cmd := exec.Command("go", "install", "-race", path)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOBIN=/tmp")
	race_cmd.Env = cmd.Env
	race_out, race_err := race_cmd.CombinedOutput()
	// Rename the race binary.
	os.Rename(GetAutoRoutePath(false), GetAutoRoutePath(true))
	out, err := cmd.CombinedOutput()
	ret_err := race_err
	ret_out := race_out
	if err != nil {
		ret_err = err
		ret_out = out
	}
	return string(ret_out), ret_err
}

func init() {
	out, err := BuildBinary("github.com/AutoRoute/node/autoroute")
	if err != nil {
		log.Fatalf("%s:%s:", err, out)
	}
}

// Gets the path for the version of the autoroute binary.
// Args:
//  race: Whether to use the version with race checking.
// Returns:
//  The path requested.
func GetAutoRoutePath(race bool) string {
	path := "/tmp/autoroute"
	if race {
	  path += "_race"
	}

	return path
}
