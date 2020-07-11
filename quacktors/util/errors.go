package util

import "errors"

func PidDoesNotExistError() error {
	return errors.New("no pid was created for this goid")
}
