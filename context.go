package quacktors

import (
	"github.com/rs/zerolog/log"
	"reflect"
)

type Context struct {
	self *Pid
}

func (c *Context) Self() *Pid {
	return c.self
}

func (c *Context) Send(to *Pid, message Message) {
	t := reflect.ValueOf(message).Type().Kind()

	if t == reflect.Ptr {
		panic("Send cannot be called with a pointer to a Message")
	}

	doSend(to, message)
}

func (c *Context) Kill(pid *Pid) {
	go func() {
		if pid.MachineId != machineId {
			log.Debug().
				Str("target_pid", pid.String()).
				Str("machine_id", pid.MachineId).
				Msg("pid to kill is not on this machine, forwarding to remote machine")

			m, ok := getMachine(pid.MachineId)

			if ok {
				m.quitChan <- pid
				return
			}

			log.Warn().
				Str("target_pid", pid.String()).
				Str("machine_id", pid.MachineId).
				Msg("remote machine is not registered, couldn't kill pid")

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

	log.Info().
		Str("monitored_pid", pid.String()).
		Str("monitor_pid", c.self.String()).
		Msg("setting up monitor")

	go func() {
		if pid.MachineId != machineId {
			log.Debug().
				Str("monitored_pid", pid.String()).
				Str("monitor_pid", c.self.String()).
				Str("machine_id", pid.MachineId).
				Msg("pid to monitor is not on this machine, forwarding to remote machine")

			m, ok := getMachine(pid.MachineId)
			if ok {
				okChan <- true

				m.monitorChan <- remoteMonitorTuple{From: c.self, To: pid}
				return
			}

			log.Warn().
				Str("monitored_pid", pid.String()).
				Str("monitor_pid", c.self.String()).
				Str("machine_id", pid.MachineId).
				Msg("remote machine is not registered, couldn't monitor pid")

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

		log.Warn().
			Str("monitored_pid", pid.String()).
			Str("monitor_pid", c.self.String()).
			Msg("monitored pid is either dead or on a machine that disconnected, sending out DownMessage to monitor immediately")

		doSend(c.self, DownMessage{Who: pid})
		return &NoopAbortable{}
	}
}
