package quacktors

type Abortable interface {
	Abort()
}

type MonitorAbortable struct {
	pid  *Pid
	self *Pid
}

func (ma *MonitorAbortable) Abort() {
	go func() {
		if ma.pid.MachineId != machineId {
			//Monitor is not on this machine

			m, ok := getMachine(ma.pid.MachineId)

			if ok {
				//send demonitor request to demonitor channel on the machine connection
				m.demonitorChan <- remoteMonitorTuple{from: ma.self, to: ma.pid}
				return
			}

			return
		}

		ma.pid.demonitorChanMu.RLock()
		defer ma.pid.demonitorChanMu.RUnlock()

		if ma.pid.demonitorChan == nil {
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
