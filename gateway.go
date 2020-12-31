package quacktors

import (
	"errors"
	"net"
)

/*
When connecting to a remote machine, quacktors works with two TCP streams
One for messages and another one for system commands (monitor, demonitor, kill)
*/

func startMessageGateway() (uint16, error) {
	return startServer(func(portChan chan int, errorChan chan error) {
		listener, err := net.Listen("tcp", ":0")

		if err != nil {
			errorChan <- errors.New("couldn't start message gateway on random port")
			return
		}

		port := listener.Addr().(*net.TCPAddr).Port
		portChan <- port

		return
	})
}

func startGeneralPurposeGateway() (uint16, error) {
	return startServer(func(portChan chan int, errorChan chan error) {
		listener, err := net.Listen("tcp", ":0")

		if err != nil {
			errorChan <- errors.New("couldn't start general purpose gateway on random port")
			return
		}

		port := listener.Addr().(*net.TCPAddr).Port
		portChan <- port

		return
	})
}

func startServer(callback func(chan int, chan error)) (uint16, error) {
	portChan := make(chan int)
	errChan := make(chan error)

	go callback(portChan, errChan)

	select {
	case p := <-portChan:
		return uint16(p), nil
	case err := <-errChan:
		return 0, err
	}
}
