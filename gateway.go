package quacktors

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/Azer0s/qpmd"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"net"
)

/*
When connecting to a remote machine, quacktors works with two TCP streams
One for messages and another one for system commands (monitor, demonitor, kill)
*/

func startMessageGateway() (uint16, error) {
	return startServer(func(portChan chan int, errorChan chan error) {
		logger.Info("starting message gateway")

		listener, err := net.Listen("tcp", ":0")

		if err != nil {
			errorChan <- errors.New("couldn't start message gateway on random port")
			return
		}

		port := listener.Addr().(*net.TCPAddr).Port

		logger.Debug("started message gatway",
			"port", port)

		portChan <- port

		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Warn("there was an error while accepting new connection to message gateway",
					"error", err)
				_ = conn.Close()
				continue
			}

			go handleMessageClient(conn)
		}
	})
}

func handleMessageClient(conn net.Conn) {
	c := conn.RemoteAddr().String()

	defer func() {
		logger.Info("closing connection to message gateway",
			"client", c)
		err := conn.Close()
		if err != nil {
			logger.Warn("there was an error while closing connection to the message gateway",
				"client", c,
				"error", err)
			return
		}
	}()

	logger.Info("handling new message gateway connection from remote machine",
		"client", c)

	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if n == 0 || err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("remote machine disconnected from message gateway",
					"client", c)
			} else {
				logger.Warn("there was an error while reading incoming message from remote machine",
					"client", c,
					"error", err)
			}
			return
		}

		msgData := make(map[string]interface{})

		err = msgpack.Unmarshal(buf[:n], &msgData)
		if err != nil {
			logger.Warn("there was an error while unmarshalling incoming message from remote machine",
				"client", c,
				"error", err)
			return
		}

		go func(data map[string]interface{}) {
			pidId := data[toVal].(string)
			toPid, ok := getByPidId(pidId)

			logger.Trace("received new message from remote machine for pid on local system",
				"client", c,
				"pid_id", pidId)

			if !ok {
				logger.Warn("couldn't find pid id target of remote message on local system",
					"client", c,
					"pid_id", pidId)
				return
			}

			byteBuf := bytes.NewBuffer(data[messageVal].([]byte))
			dec := gob.NewDecoder(byteBuf)

			var msg Message

			err = dec.Decode(&msg)

			if err != nil {
				logger.Warn("there was an error while decoding incoming message from remote machine",
					"client", c,
					"pid_id", pidId)
				return
			}

			if d, ok := msg.(DownMessage); ok {
				//if we receive a DownMessage, we can remove the link from the remote connection to the monitor

				m, ok := getMachine(d.Who.MachineId)

				if ok && m.connected {
					m.removeRemoteMonitor(remoteMonitorTuple{
						From: toPid,
						To:   d.Who,
					})
				}
			}

			doSend(toPid, msg)
		}(msgData)
	}
}

func startGeneralPurposeGateway() (uint16, error) {
	return startServer(func(portChan chan int, errorChan chan error) {
		logger.Info("starting general purpose gateway")

		listener, err := net.Listen("tcp", ":0")

		if err != nil {
			errorChan <- errors.New("couldn't start general purpose gateway on random port")
			return
		}

		port := listener.Addr().(*net.TCPAddr).Port

		logger.Debug("started general purpose gatway",
			"port", port)

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
				logger.Warn("there was an error while accepting new connection to general purpose gateway",
					"error", err)
				_ = conn.Close()
				continue
			}

			go handleGpClient(conn)
		}
	})
}

func handleGpClient(conn net.Conn) {
	c := conn.RemoteAddr().String()

	defer func() {
		logger.Info("closing connection to general purpose gateway",
			"client", c)
		err := conn.Close()
		if err != nil {
			logger.Warn("there was an error while closing connection to general purpose gateway",
				"client", c,
				"error", err)
			return
		}
	}()

	logger.Info("handling new general purpose gateway connection from remote machine",
		"client", c)

	req, err := readRequest(conn)

	if err != nil {
		logger.Warn("there was an error while reading the initial hello request to the general purpose gateway",
			"client", c,
			"error", err)
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
		Data:         make(map[string]interface{}),
	})

	if err != nil {
		logger.Warn("there was an error while responding to the initial hello request to the general purpose gateway",
			"client", c,
			"error", err)
		return
	}

	//if this is a back-connect, skip right to handling requests
	//if not, propagate the machine to all connected machines
	err = propagateMachineIfNotExists(m)

	if err != nil {
		logger.Warn("there was an error while attempting to propagate new connection information to connected machines",
			"client", c,
			"error", err)
		return
	}

	defer func() {
		machine, ok := getMachine(m.MachineId)

		if ok {
			if machine.connected {
				machine.stop()
			}
		}
	}()

	defer func() {
		remoteMonitorQuitAbortablesMu.RLock()
		defer remoteMonitorQuitAbortablesMu.RUnlock()

		for _, abortable := range remoteMonitorQuitAbortables {
			abortable.Abort()
		}
	}()

	for {
		r, err := readRequest(conn)

		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("remote machine disconnected from general purpose gateway",
					"client", c)
			} else {
				logger.Warn("there was an error while reading incoming command from remote machine",
					"client", c,
					"error", err)
			}
			return
		}

		go handleGpRequest(r, c)
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
				logger.Debug("propagating new connection information to connected machine",
					"machine_id", m.MachineId)

				machine.newConnectionChan <- m
			}
		}
	}

	return nil
}

func handleGpRequest(req qpmd.Request, client string) {
	switch req.RequestType {
	case quitMessageType:
		pidId := req.Data[pidVal].(string)
		p, ok := getByPidId(pidId)

		logger.Debug("received quit command from remote machine for pid on local system",
			"client", client,
			"pid_id", pidId)

		if !ok {
			logger.Warn("couldn't find pid id target of remote kill command on local system",
				"client", client,
				"pid_id", pidId)
			return
		}

		p.die()

	case monitorMessageType:
		fromPid, err := parsePid(req.Data[fromVal].(map[string]interface{}))

		if err != nil {
			logger.Warn("there was an error while trying to decode PID data (monitor) for monitor request from remote machine",
				"client", client,
				"error", err)
			return
		}

		toPid, err := parsePid(req.Data[toVal].(map[string]interface{}))

		if err != nil {
			logger.Warn("there was an error while trying to decode PID data (monitored PID) for monitor request from remote machine",
				"client", client,
				"error", err)
			return
		}

		p, ok := getByPidId(toPid.Id)

		if !ok {
			logger.Warn("couldn't find pid id target of remote monitor request on local system",
				"client", client,
				"pid_id", toPid.Id)
			return
		}

		remoteCtx := Context{
			self:     fromPid,
			sendLock: nil,
			Logger:   contextLogger{},
			deferred: make([]func(), 0),
		}

		remoteMonitorQuitAbortablesMu.Lock()
		defer remoteMonitorQuitAbortablesMu.Unlock()

		//TODO: log remote monitor request

		remoteMonitorQuitAbortables[fromPid.String()+"_"+p.String()] = remoteCtx.Monitor(p)

	case demonitorMessageType:
		fromPid, err := parsePid(req.Data[fromVal].(map[string]interface{}))

		if err != nil {
			logger.Warn("there was an error while trying to decode PID data (monitor) for demonitor request from remote machine",
				"client", client,
				"error", err)
			return
		}

		toPid, err := parsePid(req.Data[toVal].(map[string]interface{}))

		if err != nil {
			logger.Warn("there was an error while trying to decode PID data (monitored PID) for demonitor request from remote machine",
				"client", client,
				"error", err)
			return
		}

		remoteMonitorQuitAbortablesMu.Lock()
		defer remoteMonitorQuitAbortablesMu.Unlock()

		//TODO: log remote demonitor request

		name := fromPid.String() + "_" + toPid.String()

		remoteMonitorQuitAbortables[name].Abort()

		delete(remoteMonitorQuitAbortables, name)

	case newConnectionMessageType:
		m, err := parseMachine(req.Data[machineVal].(map[string]interface{}))

		if err != nil {
			logger.Warn("there was an error while trying to decode new connection information from remote machine",
				"client", client,
				"error", err)
			return
		}

		logger.Debug("received new connection information from remote machine",
			"client", client)

		err = propagateMachineIfNotExists(m)

		if err != nil {
			logger.Warn("there was an error while attempting to propagate new connection information to connected machines",
				"client", client,
				"error", err)
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
