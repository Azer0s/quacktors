package quacktors

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/Azer0s/qpmd"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v5"
	"net"
)

const quitMessageType = "quit"
const monitorMessageType = "monitor"
const demonitorMessageType = "demonitor"
const newConnectionMessageType = "new_connection"

const fromVal = "from"
const toVal = "to"

const messageVal = "message"

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
		log.Info().
			Str("machine_id", m.MachineId).
			Msg("stopping connection to remote machine")

		m.gatewayQuitChan <- true
		m.gpQuitChan <- true

		//TODO: notify monitors

		deleteMachine(m.MachineId)
	}()
}

func startMessageClient(m *Machine, messageChan <-chan remoteMessageTuple, gatewayQuitChan <-chan bool, okChan chan<- bool, errorChan chan<- error) {
	log.Debug().
		Str("machine_id", m.MachineId).
		Msg("starting message client for remote machine")

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", m.Address, m.MessageGatewayPort))
	if err != nil {
		errorChan <- err
		return
	}

	okChan <- true

	for {
		select {
		case message := <-messageChan:
			byteBuf := new(bytes.Buffer)
			enc := gob.NewEncoder(byteBuf)

			var inter Message
			inter = message.Message

			_ = enc.Encode(&inter)

			b, err := msgpack.Marshal(map[string]interface{}{
				toVal:      message.To.Id,
				messageVal: byteBuf.Bytes(),
			})

			if err != nil {
				log.Warn().
					Str("receiver_pid", message.To.String()).
					Str("machine_id", m.MachineId).
					Err(err).
					Msg("there was an error while sending a message to a remote machine")
				m.stop()
			}

			_, err = conn.Write(b)

			if err != nil {
				log.Warn().
					Str("receiver_pid", message.To.String()).
					Str("machine_id", m.MachineId).
					Err(err).
					Msg("there was an error while sending a message to a remote machine")
				m.stop()
			}
		case <-gatewayQuitChan:
			_ = conn.Close()
			return
		}
	}
}

func startGpClient(m *Machine, gpQuitChan <-chan bool, quitChan <-chan *Pid, monitorChan <-chan remoteMonitorTuple, demonitorChan <-chan remoteMonitorTuple, newConnectionChan <-chan *Machine, okChan chan<- bool, errorChan chan<- error) {
	log.Debug().
		Str("machine_id", m.MachineId).
		Msg("starting general purpose client for remote machine")

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
					pidVal: p.Id,
				},
			})
			if err != nil {
				log.Warn().
					Str("target_pid", p.String()).
					Str("machine_id", m.MachineId).
					Err(err).
					Msg("there was an error while sending a kill command to a remote machine")
				m.stop()
			}
		case r := <-monitorChan:
			//Note: when we monitor a foreign pid, we also have to link up the remote machine
			//to the actual monitor. I.e. if the connection to the remote machine goes down, we also have to send out
			//down messages to the monitors
			err := sendRequest(conn, qpmd.Request{
				RequestType: monitorMessageType,
				Data: map[string]interface{}{
					fromVal: r.From,
					toVal:   r.To,
				},
			})
			if err != nil {
				log.Warn().
					Str("monitor", r.From.String()).
					Str("monitored_pid", r.To.String()).
					Str("machine_id", m.MachineId).
					Err(err).
					Msg("there was an error while sending a monitor request to a remote machine")
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
				log.Warn().
					Str("monitor", r.From.String()).
					Str("monitored_pid", r.To.String()).
					Str("machine_id", m.MachineId).
					Err(err).
					Msg("there was an error while sending a demonitor request to a remote machine")
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
				log.Warn().
					Str("new_machine_id", machine.MachineId).
					Str("machine_id", m.MachineId).
					Err(err).
					Msg("there was an error while sending new connection information to a remote machine")
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
	messageChan := make(chan remoteMessageTuple, 2000)
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

	log.Info().
		Str("machine_id", m.MachineId).
		Msg("connecting to a remote machine")

	go startMessageClient(m, messageChan, gatewayQuitChan, okChan, errorChan)

	select {
	case err := <-errorChan:
		log.Warn().
			Str("machine_id", m.MachineId).
			Err(err).
			Msg("there was an error while connecting to a remote machine")
		return err
	case <-okChan:
		//everything went fine
	}

	go startGpClient(m, gpQuitChan, quitChan, monitorChan, demonitorChan, newConnectionChan, okChan, errorChan)

	select {
	case err := <-errorChan:
		gatewayQuitChan <- true
		log.Warn().
			Str("machine_id", m.MachineId).
			Err(err).
			Msg("there was an error while connecting to a remote machine")
		return err
	case <-okChan:
		//everything went fine
	}

	log.Info().
		Str("machine_id", m.MachineId).
		Msg("successfully established connection to a remote machine")

	return nil
}
