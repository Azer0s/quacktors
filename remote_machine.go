package quacktors

import (
	"errors"
	"fmt"
	"github.com/Azer0s/qpmd"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"net"
)

const quitMessageType = "quit"
const monitorMessageType = "monitor"
const demonitorMessageType = "demonitor"
const newConnectionMessageType = "new_connection"

const fromVal = "from"
const toVal = "to"

const machineVal = "machine"

type Machine struct {
	MachineId          string
	Address            string
	MessageGatewayPort uint16
	gatewayQuitChan    chan bool
	GeneralPurposePort uint16
	gpQuitChan         chan bool
	quitChan           chan<- *Pid
	messageChan        chan<- remoteMessageTuple
	monitorChan        chan<- remoteMonitorTuple
	demonitorChan      chan<- remoteMonitorTuple
	newConnectionChan  chan<- *Machine
	scheduled          map[string]chan bool
}

func (m *Machine) stop() {
	go func() {
		logger.Info("stopping connection to remote machine",
			zap.String("machine_id", m.MachineId),
		)

		m.gatewayQuitChan <- true
		m.gpQuitChan <- true

		//TODO: notify monitors

		deleteMachine(m.MachineId)
	}()
}

func startMessageClient(m *Machine, messageChan <-chan remoteMessageTuple, gatewayQuitChan <-chan bool, okChan chan<- bool, errorChan chan<- error) {
	logger.Debug("starting message client for remote machine",
		zap.String("machine_id", m.MachineId),
	)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", m.Address, m.MessageGatewayPort))
	if err != nil {
		errorChan <- err
		return
	}

	okChan <- true

	for {
		select {
		case message := <-messageChan:
			b, err := msgpack.Marshal(message)

			if err != nil {
				logger.Warn("there was an error while sending a message to a remote machine",
					zap.String("receiver_pid", message.To.String()),
					zap.String("machine_id", m.MachineId),
					zap.Error(err),
				)
				m.stop()
			}

			_, err = conn.Write(b)

			if err != nil {
				logger.Warn("there was an error while sending a message to a remote machine",
					zap.String("receiver_pid", message.To.String()),
					zap.String("machine_id", m.MachineId),
					zap.Error(err),
				)
				m.stop()
			}
		case <-gatewayQuitChan:
			_ = conn.Close()
			return
		}
	}
}

func startGpClient(m *Machine, gpQuitChan <-chan bool, quitChan <-chan *Pid, monitorChan <-chan remoteMonitorTuple, demonitorChan <-chan remoteMonitorTuple, newConnectionChan <-chan *Machine, okChan chan<- bool, errorChan chan<- error) {
	logger.Debug("starting general purpose client for remote machine",
		zap.String("machine_id", m.MachineId),
	)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", m.Address, m.GeneralPurposePort))
	if err != nil {
		errorChan <- err
		return
	}

	err = sendRequest(conn, qpmd.Request{
		RequestType: qpmd.REQUEST_HELLO,
		Data: map[string]interface{}{
			qpmd.MACHINE_ID:           machineId,
			qpmd.MESSAGE_GATEWAY_PORT: messageGatewayPort,
			qpmd.GP_GATEWAY_PORT:      gpGatewayPort,
		},
	})

	if err != nil {
		errorChan <- err
		return
	}

	res, err := readResponse(conn)

	if err != nil {
		errorChan <- err
		return
	}

	if res.ResponseType != qpmd.RESPONSE_OK {
		errorChan <- errors.New("remote machine returned non okay result")
		return
	}

	okChan <- true

	for {
		select {
		case p := <-quitChan:
			err := sendRequest(conn, qpmd.Request{
				RequestType: quitMessageType,
				Data: map[string]interface{}{
					pidVal: p,
				},
			})
			if err != nil {
				logger.Warn("there was an error while sending a kill command to a remote machine",
					zap.String("target_pid", p.String()),
					zap.String("machine_id", m.MachineId),
					zap.Error(err),
				)
				m.stop()
			}
		case r := <-monitorChan:
			err := sendRequest(conn, qpmd.Request{
				RequestType: monitorMessageType,
				Data: map[string]interface{}{
					fromVal: r.From,
					toVal:   r.To,
				},
			})
			if err != nil {
				logger.Warn("there was an error while sending a monitor request to a remote machine",
					zap.String("monitor", r.From.String()),
					zap.String("monitored_pid", r.To.String()),
					zap.String("machine_id", m.MachineId),
					zap.Error(err),
				)
				m.stop()
			}
		case r := <-demonitorChan:
			err := sendRequest(conn, qpmd.Request{
				RequestType: demonitorMessageType,
				Data: map[string]interface{}{
					fromVal: r.From,
					toVal:   r.To,
				},
			})
			if err != nil {
				logger.Warn("there was an error while sending a demonitor request to a remote machine",
					zap.String("monitor", r.From.String()),
					zap.String("monitored_pid", r.To.String()),
					zap.String("machine_id", m.MachineId),
					zap.Error(err),
				)
				m.stop()
			}
		case machine := <-newConnectionChan:
			err := sendRequest(conn, qpmd.Request{
				RequestType: newConnectionMessageType,
				Data: map[string]interface{}{
					machineVal: machine,
				},
			})
			if err != nil {
				logger.Warn("there was an error while sending new connection information to a remote machine",
					zap.String("new_machine_id", machine.MachineId),
					zap.String("machine_id", m.MachineId),
					zap.Error(err),
				)
				m.stop()
			}
		case <-gpQuitChan:
			_ = conn.Close()
			return
		}
	}
}

func (m *Machine) connect() error {
	quitChan := make(chan *Pid, 100)
	messageChan := make(chan remoteMessageTuple, 100)
	monitorChan := make(chan remoteMonitorTuple, 100)
	demonitorChan := make(chan remoteMonitorTuple, 100)
	newConnectionChan := make(chan *Machine, 100)

	m.quitChan = quitChan
	m.messageChan = messageChan
	m.monitorChan = monitorChan
	m.demonitorChan = demonitorChan
	m.newConnectionChan = newConnectionChan

	//Buffer size of 1 to avoid leaks if both connections fail
	gatewayQuitChan := make(chan bool, 1)
	gpQuitChan := make(chan bool, 1)

	m.gatewayQuitChan = gatewayQuitChan
	m.gpQuitChan = gpQuitChan

	errorChan := make(chan error)
	okChan := make(chan bool)

	logger.Info("connecting to a remote machine",
		zap.String("machine_id", m.MachineId),
	)

	go startMessageClient(m, messageChan, gatewayQuitChan, okChan, errorChan)

	select {
	case err := <-errorChan:
		logger.Warn("there was an error while connecting to a remote machine",
			zap.String("machine_id", m.MachineId),
			zap.Error(err),
		)
		return err
	case <-okChan:
		//everything went fine
	}

	go startGpClient(m, gpQuitChan, quitChan, monitorChan, demonitorChan, newConnectionChan, okChan, errorChan)

	select {
	case err := <-errorChan:
		gatewayQuitChan <- true
		logger.Warn("there was an error while connecting to a remote machine",
			zap.String("machine_id", m.MachineId),
			zap.Error(err),
		)
		return err
	case <-okChan:
		//everything went fine
	}

	logger.Info("successfully established connection to a remote machine",
		zap.String("machine_id", m.MachineId),
	)

	return nil
}
