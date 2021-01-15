package quacktors

import (
	"github.com/opentracing/opentracing-go"
	"reflect"
	"sync"
	"time"
)

type Context struct {
	span      opentracing.Span
	traceFork func(ctx opentracing.SpanContext) opentracing.SpanReference
	traceName string
	self      *Pid
	sendLock  *sync.Mutex
	Logger    contextLogger
	deferred  []func()
}

//Trace enables distributed tracing for the actor
//(quacktors will create a ChildSpan with the operationName
//set to the provided name).
func (c *Context) Trace(name string) {
	if name == "" {
		panic("actor trace name cannot be empty string")
	}

	c.traceName = name
}

//TraceFork sets the default fork mechanism for
//incoming SpanContexts. By default, this is set
//to opentracing.FollowsFrom.
func (c *Context) TraceFork(traceFork func(ctx opentracing.SpanContext) opentracing.SpanReference) {
	c.traceFork = traceFork
}

func (c *Context) Span() opentracing.Span {
	return c.span
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

	var spanContext opentracing.SpanContext
	if c.span != nil {
		spanContext = c.span.Context()
	}

	doSend(to, message, spanContext)
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

		doSend(c.self, DisconnectMessage{MachineId: machine.MachineId, Address: machine.Address}, nil)
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

		doSend(c.self, DownMessage{Who: pid}, nil)
		return &noopAbortable{}
	}
}
