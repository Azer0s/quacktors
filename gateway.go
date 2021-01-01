package quacktors

import (
	"encoding/json"
	"errors"
	"github.com/Azer0s/qpmd"
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

		for {
			//As soon as we accept a connection, forward a "new_connection" request to our connected machines
			//If they don't have that connection, they should register it, connect to it and forward the information
			//To all of their connected machines
			//If they do have that connection, they should do nothing

			//Then, connect back to the requestor machine
			//The requestor will then forward our connection to their connected machines and propagate

			conn, err := listener.Accept()
			if err != nil {
				_ = conn.Close()
				continue
			}

			go handleGpClient(conn)
		}
	})
}

func handleGpClient(conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			return
		}
	}()

	req, err := readRequest(conn)

	if err != nil {
		return
	}

	ip := conn.RemoteAddr().(*net.TCPAddr).IP

	//Sometimes, go wants to force us to use IPv6, but there are some weird bugs
	//("too many colons in address"), so I force IPv4 instead
	if ip.IsLoopback() {
		ip = net.IPv4(127, 0, 0, 1)
	}

	m := &Machine{
		MachineId:          req.Data[qpmd.MACHINE_ID].(string),
		Address:            ip.String(),
		MessageGatewayPort: req.Data[qpmd.MESSAGE_GATEWAY_PORT].(uint16),
		GeneralPurposePort: req.Data[qpmd.GP_GATEWAY_PORT].(uint16),
	}

	err = sendResponse(conn, qpmd.Response{
		ResponseType: qpmd.RESPONSE_OK,
		Data:         map[string]interface{}{},
	})

	if err != nil {
		return
	}

	//if this is a back-connect, skip right to handling requests
	//if not, propagate the machine to all connected machines
	err = propagateMachineIfNotExists(m)

	if err != nil {
		return
	}

	for {
		r, err := readRequest(conn)

		if err != nil {
			return
		}

		go handleGpRequest(r)
	}
}

func propagateMachineIfNotExists(m *Machine) error {
	if _, ok := getMachine(m.MachineId); !ok {
		err := m.connect()

		if err != nil {
			return err
		}

		registerMachine(m)

		machinesMu.RLock()
		defer machinesMu.RUnlock()

		for _, machine := range machines {
			if machine.MachineId != m.MachineId {
				machine.newConnectionChan <- m
			}
		}
	}

	return nil
}

func handleGpRequest(req qpmd.Request) {
	switch req.RequestType {
	case quitMessageType:
	case monitorMessageType:
	case demonitorMessageType:
	case newConnectionMessageType:
		b, err := json.Marshal(req.Data[machineVal])
		if err != nil {
			return
		}

		m := &Machine{}
		err = json.Unmarshal(b, m)
		if err != nil {
			return
		}

		err = propagateMachineIfNotExists(m)

		if err != nil {
			return
		}
	}
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
