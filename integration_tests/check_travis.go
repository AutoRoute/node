package integration_tests

import (
	"os/user"
)

// Checks if we are running on Travis or not.
func CheckTravis() (bool, error) {
	user, err := user.Current()
	if err != nil {
		return false, err
	}
	return user.Username == "travis", nil
}
