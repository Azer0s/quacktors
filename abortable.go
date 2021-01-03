package quacktors

type Abortable interface {
	Abort()
}

type MonitorAbortable struct {
	pid  *Pid
	self *Pid
}

func (ma *MonitorAbortable) Abort() {
	logger.Debug("demonitoring pid",
		"monitored_pid", ma.pid.String(),
		"monitor_pid", ma.self.String(),
	)

	go func() {
		if ma.pid.MachineId != machineId {
			//Monitor is not on this machine

			logger.Debug("monitor to abort is not on this machine, forwarding to remote machine",
				"monitored_pid", ma.pid.String(),
				"monitor_pid", ma.self.String(),
				"machine_id", ma.pid.MachineId)

			m, ok := getMachine(ma.pid.MachineId)

			if ok && m.conntected {
				//send demonitor request to demonitor channel on the machine connection
				m.demonitorChan <- remoteMonitorTuple{From: ma.self, To: ma.pid}
				return
			}

			logger.Warn("remote machine is not registered, couldn't abort monitor",
				"monitored_pid", ma.pid.String(),
				"monitor_pid", ma.self.String(),
				"machine_id", ma.pid.MachineId)

			return
		}

		defer func() {
			if r := recover(); r != nil {
				//This happens if we write to the demonitorChan while the actor is being closed
			}
		}()

		if ma.pid.demonitorChan == nil {
			logger.Warn("pid to demonitor is already down",
				"monitored_pid", ma.pid.String(),
				"monitor_pid", ma.self.String())
			return
		}

		ma.pid.demonitorChan <- ma.self
	}()
}

type SendAfterAbortable struct {
}

func (sa *SendAfterAbortable) Abort() {

}

type NoopAbortable struct {
}

func (na *NoopAbortable) Abort() {

}
