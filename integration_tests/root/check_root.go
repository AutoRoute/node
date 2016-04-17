package root

import (
	"errors"
	"fmt"
	"os/user"
)

func CheckRoot() error {
	user, err := user.Current()
	if err != nil {
		return errors.New(
			"Unable to determine user.. These tests require root and may fail.")
	}
	if user.Username != "root" {
		return fmt.Errorf(
			"Current user is %s which is not root. These tests require root.",
			user.Username)
	}
	return nil
}
