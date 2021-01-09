package quacktors

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/Azer0s/qpmd"
	"github.com/vmihailenco/msgpack/v5"
	"net"
	"sync"
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
	connected          bool
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
	//Stores channels to scheduled monitors
	scheduled map[string]chan bool
	//Stores channels to tell a monitor task to quit (when a pid is demonitored)
	monitorQuitChannels map[string]chan bool
	monitorsMu          *sync.Mutex
}

func (m *Machine) stop() {
	go func() {
		m.connected = false

		logger.Info("stopping connections to remote machine",
			"machine_id", m.MachineId)

		deleteMachine(m.MachineId)

		m.gatewayQuitChan <- true
		m.gpQuitChan <- true

		close(m.messageChan)

		m.monitorsMu.Lock()
		defer m.monitorsMu.Unlock()

		if len(m.scheduled) != 0 {
			//Terminate all scheduled events/send down message to monitor tasks
			logger.Debug("sending out scheduled events after remote machine disconnect",
				"machine_id", m.MachineId)

			for k, ch := range m.scheduled {
				ch <- true
				close(ch)
				delete(m.scheduled, k)
			}
		}

		if len(m.monitorQuitChannels) != 0 {
			logger.Debug("deleting machine connection monitor abort channels",
				"machine_id", m.MachineId)

			//Delete monitorQuitChannels
			for n, c := range m.monitorQuitChannels {
				close(c)
				delete(m.monitorQuitChannels, n)
			}
		}

		m.monitorQuitChannels = nil
	}()
}

func (m *Machine) setupMonitor(monitor *Pid) {
	name := monitor.String()

	monitorChannel := make(chan bool)
	m.scheduled[name] = monitorChannel

	monitorQuitChannel := make(chan bool)
	m.monitorQuitChannels[name] = monitorQuitChannel

	go func() {
		select {
		case <-monitorQuitChannel:
			return
		case <-monitorChannel:
			doSend(monitor, DisconnectMessage{MachineId: m.MachineId, Address: m.Address})
		}
	}()
}

func (m *Machine) setupRemoteMonitor(r remoteMonitorTuple) {
	m.monitorsMu.Lock()
	defer m.monitorsMu.Unlock()

	name := r.From.String() + "_" + r.To.String()

	monitorChannel := make(chan bool)
	m.scheduled[name] = monitorChannel

	monitorQuitChannel := make(chan bool)
	m.monitorQuitChannels[name] = monitorQuitChannel

	go func() {
		select {
		case <-monitorQuitChannel:
			return
		case <-monitorChannel:
			doSend(r.From, DownMessage{Who: r.To})
		}
	}()
}

func (m *Machine) removeRemoteMonitor(r remoteMonitorTuple) {
	m.monitorsMu.Lock()
	defer m.monitorsMu.Unlock()

	name := r.From.String() + "_" + r.To.String()

	m.monitorQuitChannels[name] <- true

	delete(m.scheduled, name)
	delete(m.monitorQuitChannels, name)
}

func (m *Machine) startMessageClient(messageChan <-chan remoteMessageTuple, gatewayQuitChan <-chan bool, okChan chan<- bool, errorChan chan<- error) {
	logger.Debug("starting message client for remote machine",
		"machine_id", m.MachineId)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", m.Address, m.MessageGatewayPort))
	if err != nil {
		errorChan <- err
		return
	}

	okChan <- true

	for {
		select {
		case message := <-messageChan:
			if d, ok := message.Message.(DownMessage); ok {
				logger.Trace("cleaning up old remote monitor abortable, local PID just went down",
					"monitor_pid", message.To.String(),
					"monitored_pid", d.Who.String())

				remoteMonitorQuitAbortablesMu.Lock()
				delete(remoteMonitorQuitAbortables, message.To.String()+"_"+d.Who.String())
				remoteMonitorQuitAbortablesMu.Unlock()
			}

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
				logger.Warn("there was an error while sending message to remote machine",
					"receiver_pid", message.To.String(),
					"machine_id", m.MachineId,
					"error", err)
				m.stop()
			}

			_, err = conn.Write(b)

			if err != nil {
				logger.Warn("there was an error while sending message to remote machine",
					"receiver_pid", message.To.String(),
					"machine_id", m.MachineId,
					"error", err)
				m.stop()
			}
		case <-gatewayQuitChan:
			logger.Info("closing connection to remote message gateway",
				"machine_id", m.MachineId)
			_ = conn.Close()
			return
		}
	}
}

func (m *Machine) startGpClient(gpQuitChan <-chan bool, quitChan <-chan *Pid, monitorChan <-chan remoteMonitorTuple, demonitorChan <-chan remoteMonitorTuple, newConnectionChan <-chan *Machine, okChan chan<- bool, errorChan chan<- error) {
	logger.Debug("starting general purpose client for remote machine",
		"machine_id", m.MachineId)

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
				logger.Warn("there was an error while sending kill command to remote machine",
					"target_pid", p.String(),
					"machine_id", m.MachineId,
					"error", err)
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
				logger.Warn("there was an error while sending monitor request to remote machine",
					"monitor", r.From.String(),
					"monitored_pid", r.To.String(),
					"machine_id", m.MachineId,
					"error", err)
				m.stop()
			}

			//this is the above mentioned link; if the remote connection goes down, a DownMessage is sent
			//to the monitoring PID (but obviously from the local machine because the remote one already
			//disconnected)
			m.setupRemoteMonitor(r)

		case r := <-demonitorChan:
			err := sendRequest(conn, qpmd.Request{
				RequestType: demonitorMessageType,
				Data: map[string]interface{}{
					fromVal: r.From,
					toVal:   r.To,
				},
			})

			if err != nil {
				logger.Warn("there was an error while sending demonitor request to remote machine",
					"monitor", r.From.String(),
					"monitored_pid", r.To.String(),
					"machine_id", m.MachineId,
					"error", err)
				m.stop()
			}

			//remove "link" to the connection
			m.removeRemoteMonitor(r)

		case machine := <-newConnectionChan:
			err := sendRequest(conn, qpmd.Request{
				RequestType: newConnectionMessageType,
				Data: map[string]interface{}{
					machineVal: machine,
				},
			})
			if err != nil {
				logger.Warn("there was an error while sending new connection information to remote machine",
					"new_machine_id", machine.MachineId,
					"machine_id", m.MachineId,
					"error", err)
				m.stop()
			}
		case <-gpQuitChan:
			logger.Info("closing connection to remote general purpose gateway",
				"machine_id", m.MachineId)
			_ = conn.Close()
			return
		}
	}
}

func (m *Machine) connect() error {
	//quitChan, monitorChan, demonitorChan and newConnectionChan each have buffers of 100
	//this is a, sort of, "close protection" for when a remote machine disconnects

	//there is a short time frame (i.e. a couple ns) where the *Machine is closing
	//but is not yet marked "closed" or deleted from the machine register
	//in a couple ns, nothing much can happen really (if there are a lot of actors open,
	//maybe 1 or 2 will manage to send out some commands to the dangling *Machine)
	//this buffer is there as a precaution to avoid leaking goroutines (because as soon as the
	//machine is dereferenced, the channels get garbage collected and the machine has to be
	//dereferenced at some point because it's then only attached to the RemoteSystem instance which
	//will return an error once used again, forcing the application to reconnect and destroying
	//the old Machine ptr)

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

	m.scheduled = make(map[string]chan bool)
	m.monitorQuitChannels = make(map[string]chan bool)
	m.monitorsMu = &sync.Mutex{}

	//Buffer size of 2 to avoid leaks if both connections fail
	gatewayQuitChan := make(chan bool, 2)
	gpQuitChan := make(chan bool, 2)

	m.gatewayQuitChan = gatewayQuitChan
	m.gpQuitChan = gpQuitChan

	errorChan := make(chan error)
	okChan := make(chan bool)

	logger.Info("connecting to remote machine",
		"machine_id", m.MachineId)

	go m.startMessageClient(messageChan, gatewayQuitChan, okChan, errorChan)

	select {
	case err := <-errorChan:
		logger.Warn("there was an error while connecting to remote machine",
			"machine_id", m.MachineId,
			"error", err)
		return err
	case <-okChan:
		//everything went fine
	}

	go m.startGpClient(gpQuitChan, quitChan, monitorChan, demonitorChan, newConnectionChan, okChan, errorChan)

	select {
	case err := <-errorChan:
		gatewayQuitChan <- true
		logger.Warn("there was an error while connecting to remote machine",
			"machine_id", m.MachineId,
			"error", err)
		return err
	case <-okChan:
		//everything went fine
	}

	logger.Info("successfully established connection to remote machine",
		"machine_id", m.MachineId)

	m.connected = true

	return nil
}
