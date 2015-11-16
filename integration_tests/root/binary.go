package root

import (
	integration "github.com/AutoRoute/node/integration_tests"

	"log"
)

func init() {
	out, err := integration.BuildBinary("github.com/AutoRoute/loopback2")
	if err != nil {
		log.Fatalf("%s:%s:", err, out)
	}
}

func GetLoopBack2Path() string {
	return "/tmp/loopback2"
}
