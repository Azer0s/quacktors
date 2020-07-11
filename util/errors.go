package util

import "errors"

// PidDoesNotExistError returns an error message for the case a PID (i.e. a quacktor actor) does not exist for a given Goroutine ID
func PidDoesNotExistError() error {
	return errors.New("no pid was created for this goid")
}
