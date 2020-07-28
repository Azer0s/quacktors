package util

import "errors"

// PidDoesNotExistError returns an error message for the case a PID (i.e. a quacktor actor) does not exist for a given Goroutine ID
func PidDoesNotExistError() error {
	return errors.New("no pid was created for this goid")
}

// NoSuchPidInSystemError returns an error message for the case a PID (i.e. a quacktor actor) does not exist in a quacktor system
func NoSuchPidInSystemError() error {
	return errors.New("pid does not exist in system")
}

// InvalidAddressError returns an error message for the case a quacktor connection string is invalid
func InvalidAddressError() error {
	return errors.New("quacktor connection string is invalid")
}

// SystemDoesNotExistError returns an error message for the case a quacktor system does not exist in the context
func SystemDoesNotExistError() error {
	return errors.New("system does not exist")
}

// RemoteError returns an error message for the case a remote system reported an error while connecting
func RemoteConnectError() error {
	return errors.New("the remote system reported an error while connecting")
}