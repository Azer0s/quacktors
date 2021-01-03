package quacktors

import (
	"reflect"
	"sync"
)

type Context struct {
	self     *Pid
	sendLock *sync.Mutex
	Logger   contextLogger
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

func (c *Context) Kill(pid *Pid) {
	go func() {
		if pid.MachineId != machineId {
			logger.Debug("pid to kill is not on this machine, forwarding to remote machine",
				"target_pid", pid.String(),
				"machine_id", pid.MachineId)

			m, ok := getMachine(pid.MachineId)

			if ok {
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
	panic("Bye cruel world!")
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
			if ok {
				okChan <- true

				m.monitorChan <- remoteMonitorTuple{From: c.self, To: pid}
				return
			}

			logger.Warn("remote machine is not registered, couldn't monitor pid",
				"monitored_pid", pid.String(),
				"monitor_pid", c.self.String(),
				"machine_id", pid.MachineId)

			errorChan <- true
		} else {
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
		}
	}()

	select {
	case <-okChan:
		return &MonitorAbortable{
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
		return &NoopAbortable{}
	}
}
