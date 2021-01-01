package quacktors

import "go.uber.org/zap"

type Abortable interface {
	Abort()
}

type MonitorAbortable struct {
	pid  *Pid
	self *Pid
}

func (ma *MonitorAbortable) Abort() {
	logger.Debug("demonitoring pid",
		zap.String("monitored_pid", ma.pid.String()),
		zap.String("monitor_pid", ma.self.String()),
	)

	go func() {
		if ma.pid.MachineId != machineId {
			//Monitor is not on this machine

			logger.Debug("monitor to abort is not on this machine, forwarding to remote machine",
				zap.String("monitored_pid", ma.pid.String()),
				zap.String("monitor_pid", ma.self.String()),
				zap.String("machine_id", ma.pid.MachineId),
			)

			m, ok := getMachine(ma.pid.MachineId)

			if ok {
				//send demonitor request to demonitor channel on the machine connection
				m.demonitorChan <- remoteMonitorTuple{From: ma.self, To: ma.pid}
				return
			}

			logger.Warn("remote machine is not registered, couldn't abort monitor",
				zap.String("monitored_pid", ma.pid.String()),
				zap.String("monitor_pid", ma.self.String()),
				zap.String("machine_id", ma.pid.MachineId),
			)

			return
		}

		ma.pid.demonitorChanMu.RLock()
		defer ma.pid.demonitorChanMu.RUnlock()

		if ma.pid.demonitorChan == nil {
			logger.Warn("pid to demonitor is already down",
				zap.String("monitored_pid", ma.pid.String()),
				zap.String("monitor_pid", ma.self.String()),
			)
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
