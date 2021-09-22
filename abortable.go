package quacktors

//The Abortable interface defines the methods a struct has
//to implement so it can be returned by an action that can be canceled.
//It is very similar to context.Context with the key difference that
//an Abortable can only Abort and doesn't carry any further
//details about the underlying action.
type Abortable interface {
	//The Abort function aborts the underlying task (e.g. a Monitor) when called.
	Abort()
}

type monitorAbortable struct {
	pid  *Pid
	self *Pid
}

func (ma *monitorAbortable) Abort() {
	logger.Debug("demonitoring pid",
		"monitored_gpid", ma.pid.String(),
		"monitor_pid", ma.self.Id)

	go func() {
		if ma.pid.MachineId != machineId {
			//Monitor is not on this machine

			logger.Debug("monitor to abort is not on this machine, forwarding to remote machine",
				"monitored_gpid", ma.pid.String(),
				"monitor_pid", ma.self.Id,
				"machine_id", ma.pid.MachineId)

			m, ok := getMachine(ma.pid.MachineId)

			if ok && m.connected {
				//send demonitor request to demonitor channel on the machine connection
				m.demonitorChan <- remoteMonitorTuple{From: ma.self, To: ma.pid}
				return
			}

			logger.Warn("remote machine is not registered, couldn't abort monitor",
				"monitored_gpid", ma.pid.String(),
				"monitor_pid", ma.self.Id,
				"machine_id", ma.pid.MachineId)

			return
		}

		defer func() {
			if r := recover(); r != nil {
				//This happens if we write to the demonitorChan while the actor is being closed
			}
		}()

		if ma.pid.controlChan == nil {
			logger.Warn("pid to demonitor is already down",
				"monitored_gpid", ma.pid.String(),
				"monitor_pid", ma.self.Id)
			return
		}

		ma.pid.controlChan <- demonitorControlMessage{Who: ma.self}
	}()
}

type machineConnectionMonitorAbortable struct {
	machine *Machine
	monitor *Pid
}

func (ma *machineConnectionMonitorAbortable) Abort() {
	logger.Debug("demonitoring machine connection",
		"machine_id", ma.machine.MachineId,
		"monitor_pid", ma.monitor.Id)

	go func() {
		ma.machine.monitorsMu.Lock()
		defer ma.machine.monitorsMu.Unlock()

		defer func() {
			if r := recover(); r != nil {
				//This happens if we write to the demonitorChan while the actor is being closed
			}
		}()

		if !ma.machine.connected {
			logger.Warn("machine connection to demonitor is already down",
				"machine_id", ma.machine.MachineId,
				"monitor_pid", ma.monitor.Id)
			return
		}

		ma.machine.monitorQuitChannels[ma.monitor.String()] <- true
	}()
}

type sendAfterAbortable struct {
	quitChan chan bool
}

func (sa *sendAfterAbortable) Abort() {
	defer func() {
		//this can happen if the channel is already closed
		recover()
	}()

	sa.quitChan <- true
}

type noopAbortable struct {
}

func (na *noopAbortable) Abort() {

}
