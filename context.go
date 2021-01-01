package quacktors

import "go.uber.org/zap"

type Context struct {
	self *Pid
}

func (c *Context) Self() *Pid {
	return c.self
}

func (c *Context) Send(to *Pid, message Message) {
	doSend(to, message)
}

func (c *Context) Kill(pid *Pid) {
	go func() {
		if pid.MachineId != machineId {
			logger.Debug("pid to kill is not on this machine, forwarding to remote machine",
				zap.String("target_pid", pid.String()),
				zap.String("machine_id", pid.MachineId),
			)

			m, ok := getMachine(pid.MachineId)

			if ok {
				m.quitChan <- pid
				return
			}

			logger.Warn("remote machine is not registered, couldn't kill pid",
				zap.String("target_pid", pid.String()),
				zap.String("machine_id", pid.MachineId),
			)

			return
		}

		pid.quitChanMu.RLock()
		defer pid.quitChanMu.RUnlock()

		if pid.quitChan == nil {
			return
		}

		pid.quitChan <- true
	}()
}

func (c *Context) Quit() {
	panic("Bye cruel world!")
}

func (c *Context) Monitor(pid *Pid) Abortable {
	errorChan := make(chan bool)
	okChan := make(chan bool)

	go func() {
		if pid.MachineId != machineId {
			logger.Debug("pid to monitor is not on this machine, forwarding to remote machine",
				zap.String("monitored_pid", pid.String()),
				zap.String("monitor_pid", c.self.String()),
				zap.String("machine_id", pid.MachineId),
			)

			m, ok := getMachine(pid.MachineId)
			if ok {
				okChan <- true

				m.monitorChan <- remoteMonitorTuple{From: c.self, To: pid}
				return
			}

			logger.Warn("remote machine is not registered, couldn't monitor pid",
				zap.String("monitored_pid", pid.String()),
				zap.String("monitor_pid", c.self.String()),
				zap.String("machine_id", pid.MachineId),
			)

			errorChan <- true
		} else {
			pid.monitorChanMu.RLock()
			defer pid.monitorChanMu.RUnlock()

			if pid.monitorChan == nil {
				errorChan <- true
				return
			}

			okChan <- true

			pid.monitorChan <- c.self
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
		doSend(c.self, &DownMessage{Who: pid})
		return &NoopAbortable{}
	}
}
