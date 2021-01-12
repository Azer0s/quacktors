package quacktors

import (
	"reflect"
	"sync"
	"time"
)

type Context struct {
	self     *Pid
	sendLock *sync.Mutex
	Logger   contextLogger
	deferred []func()
}

func (c *Context) Defer(action func()) {
	c.deferred = append(c.deferred, action)
}

func (c *Context) Self() *Pid {
	return c.self
}

func (c *Context) Send(to *Pid, message Message) {
	t := reflect.ValueOf(message).Type().Kind()

	if t == reflect.Ptr {
		panic("Send cannot be called with a pointer to a Message")
	}

	c.sendLock.Lock()
	defer c.sendLock.Unlock()

	doSend(to, message)
}

func (c *Context) SendAfter(to *Pid, message Message, duration time.Duration) Abortable {
	quitChan := make(chan bool)

	go func() {
		defer close(quitChan)

		select {
		case <-time.After(duration):
			c.Send(to, message)
			return
		case <-quitChan:
			return
		}
	}()

	return &sendAfterAbortable{quitChan: quitChan}
}

func (c *Context) Kill(pid *Pid) {
	go func() {
		if pid.MachineId != machineId {
			logger.Debug("pid to kill is not on this machine, forwarding to remote machine",
				"target_pid", pid.String(),
				"machine_id", pid.MachineId)

			m, ok := getMachine(pid.MachineId)

			if ok && m.connected {
				m.quitChan <- pid
				return
			}

			logger.Warn("remote machine is not registered, couldn't kill pid",
				"target_pid", pid.String(),
				"machine_id", pid.MachineId)

			return
		}

		pid.die()
	}()
}

func (c *Context) Quit() {
	panic(quitAction{})
}

func (c *Context) MonitorMachine(machine *Machine) Abortable {
	machine.monitorsMu.Lock()
	defer machine.monitorsMu.Unlock()

	logger.Info("setting up machine connection monitor",
		"monitored_machine", machine.MachineId,
		"monitor_pid", c.self.String())

	if !machine.connected {
		//The remote machine already disconnected, send a down message immediately

		logger.Warn("monitored machine already disconnected, sending out DisconnectMessage to monitor immediately",
			"monitored_machine", machine.MachineId,
			"monitor_pid", c.self.String())

		doSend(c.self, DisconnectMessage{MachineId: machine.MachineId, Address: machine.Address})
		return &noopAbortable{}
	}

	machine.setupMonitor(c.self)

	return &machineConnectionMonitorAbortable{
		machine: machine,
		monitor: c.self,
	}
}

func (c *Context) Monitor(pid *Pid) Abortable {
	errorChan := make(chan bool)
	okChan := make(chan bool)

	logger.Info("setting up monitor",
		"monitored_pid", pid.String(),
		"monitor_pid", c.self.String())

	go func() {
		if pid.MachineId != machineId {
			logger.Debug("pid to monitor is not on this machine, forwarding to remote machine",
				"monitored_pid", pid.String(),
				"monitor_pid", c.self.String(),
				"machine_id", pid.MachineId)

			m, ok := getMachine(pid.MachineId)
			if ok && m.connected {
				okChan <- true

				m.monitorChan <- remoteMonitorTuple{From: c.self, To: pid}
				return
			}

			logger.Warn("remote machine is not registered, couldn't monitor pid",
				"monitored_pid", pid.String(),
				"monitor_pid", c.self.String(),
				"machine_id", pid.MachineId)

			errorChan <- true

			return
		}

		defer func() {
			if r := recover(); r != nil {
				//This happens if we write to the monitorChan while the actor is being closed
				errorChan <- true
			}
		}()

		if pid.monitorChan == nil {
			errorChan <- true
			return
		}

		pid.monitorChan <- c.self

		okChan <- true
	}()

	select {
	case <-okChan:
		return &monitorAbortable{
			pid:  pid,
			self: c.self,
		}
	case <-errorChan:
		//Either the remote machine disconnected or the actor is already dead.
		//Either way, send a down message

		logger.Warn("monitored pid is either dead or on a machine that disconnected, sending out DownMessage to monitor immediately",
			"monitored_pid", pid.String(),
			"monitor_pid", c.self.String())

		doSend(c.self, DownMessage{Who: pid})
		return &noopAbortable{}
	}
}
