package root

import (
	"os/user"
	"testing"
)

func WarnRoot(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Log(
			"Unable to determine user.. These tests require root and may fail.")
	}
	if user.Username != "root" {
		t.Logf(
			"Current user is %s which is not root. These tests require root.",
			user.Username)
	}
}
