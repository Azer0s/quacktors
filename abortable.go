package quacktors

import (
	"github.com/rs/zerolog/log"
)

type Abortable interface {
	Abort()
}

type MonitorAbortable struct {
	pid  *Pid
	self *Pid
}

func (ma *MonitorAbortable) Abort() {
	log.Debug().
		Str("monitored_pid", ma.pid.String()).
		Str("monitor_pid", ma.self.String()).
		Msg("demonitoring pid")

	go func() {
		if ma.pid.MachineId != machineId {
			//Monitor is not on this machine

			log.Debug().
				Str("monitored_pid", ma.pid.String()).
				Str("monitor_pid", ma.self.String()).
				Str("machine_id", ma.pid.MachineId).
				Msg("monitor to abort is not on this machine, forwarding to remote machine")

			m, ok := getMachine(ma.pid.MachineId)

			if ok {
				//send demonitor request to demonitor channel on the machine connection
				m.demonitorChan <- remoteMonitorTuple{From: ma.self, To: ma.pid}
				return
			}

			log.Warn().
				Str("monitored_pid", ma.pid.String()).
				Str("monitor_pid", ma.self.String()).
				Str("machine_id", ma.pid.MachineId).
				Msg("remote machine is not registered, couldn't abort monitor")

			return
		}

		ma.pid.demonitorChanMu.RLock()
		defer ma.pid.demonitorChanMu.RUnlock()

		if ma.pid.demonitorChan == nil {
			log.Warn().
				Str("monitored_pid", ma.pid.String()).
				Str("monitor_pid", ma.self.String()).
				Msg("pid to demonitor is already down")
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
