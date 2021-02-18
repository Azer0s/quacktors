package quacktors

import (
	"github.com/opentracing/opentracing-go"
	"reflect"
	"sync"
	"time"
)

//The Context struct defines the actor context and
//provides ways for an actor to interact with the
//rest of the system. Actors are provided a
//Context instance on Init and Run. Actors should
//only use the provided context to interact with
//other actors as the Context also stores things
//like current Span or a pointer to the acto
//specific send mutex.
type Context struct {
	span                  opentracing.Span
	traceFork             func(ctx opentracing.SpanContext) opentracing.SpanReference
	traceName             string
	self                  *Pid
	sendLock              *sync.Mutex
	Logger                contextLogger
	deferred              []func()
	passthroughPoisonPill bool
}

//PassthroughPoisonPill enables message passthrough for
//PoisonPill messages. If set to true, PoisonPill messages
//will not shut down the actor but be forwarded to the handler
//function.
func (c *Context) PassthroughPoisonPill(val bool) {
	c.passthroughPoisonPill = val
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

//Span returns the current opentracing.Span. This will
//always be nil unless Trace was called with a service
//name in the Init function of the actor.
func (c *Context) Span() opentracing.Span {
	return c.span
}

//Defer defers an action to after an actor has gone down.
//The same general advice applies to the Defer function
//as to the built-in Go defer (e.g. avoid defers in
//for loops, no nil function defers, etc). Deferred
//actor functions should not panic (because nothing will
//happen if they do, quacktors just recovers the panic).
func (c *Context) Defer(action func()) {
	c.deferred = append(c.deferred, action)
}

//Self returns the PID of the calling actor.
func (c *Context) Self() *Pid {
	return c.self
}

//Send sends a Message to another actor by its PID.
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

//SendAfter schedules a Message to be sent to another
//actor by its PID after a timer has finished. SendAfter
//also returns an Abortable so the scheduled Send can
//be stopped. If the sending actor goes down before the
//timer has completed, the Send operation is still executed.
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

//Kill kills another actor by its PID.
func (c *Context) Kill(pid *Pid) {
	go func() {
		if pid.MachineId != machineId {
			logger.Debug("pid to kill is not on this machine, forwarding to remote machine",
				"target_gpid", pid.String(),
				"machine_id", pid.MachineId)

			m, ok := getMachine(pid.MachineId)

			if ok && m.connected {
				m.quitChan <- pid
				return
			}

			logger.Warn("remote machine is not registered, couldn't kill pid",
				"target_gpid", pid.String(),
				"machine_id", pid.MachineId)

			return
		}

		pid.die()
	}()
}

//Quit kills the calling actor.
func (c *Context) Quit() {
	panic(quitAction{})
}

//MonitorMachine starts a monitor on a connection to
//a remote machine. As soon as the remote disconnects,
//a DisconnectMessage is sent to the monitoring actor.
//MonitorMachine also returns an Abortable so the
//monitor can be canceled (i.e. no DisconnectMessage
//will be sent out if the monitored actor goes down).
func (c *Context) MonitorMachine(machine *Machine) Abortable {
	machine.monitorsMu.Lock()
	defer machine.monitorsMu.Unlock()

	logger.Info("setting up machine connection monitor",
		"monitored_machine", machine.MachineId,
		"monitor_pid", c.self.Id)

	if !machine.connected {
		//The remote machine already disconnected, send a down message immediately

		logger.Warn("monitored machine already disconnected, sending out DisconnectMessage to monitor immediately",
			"monitored_machine", machine.MachineId,
			"monitor_pid", c.self.Id)

		doSend(c.self, DisconnectMessage{MachineId: machine.MachineId, Address: machine.Address}, nil)
		return &noopAbortable{}
	}

	machine.setupMonitor(c.self)

	return &machineConnectionMonitorAbortable{
		machine: machine,
		monitor: c.self,
	}
}

//Monitor starts a monitor on another actor. As soon as
//the actor goes down, a DownMessage is sent to the
//monitoring actor. Monitor also returns an Abortable
//so the monitor can be canceled (i.e. no DownMessage
//will be sent out if the monitored actor goes down).
func (c *Context) Monitor(pid *Pid) Abortable {
	errorChan := make(chan bool)
	okChan := make(chan bool)

	logger.Info("setting up monitor",
		"monitored_gpid", pid.String(),
		"monitor_pid", c.self.Id)

	go func() {
		if pid.MachineId != machineId {
			logger.Debug("pid to monitor is not on this machine, forwarding to remote machine",
				"monitored_gpid", pid.String(),
				"monitor_pid", c.self.Id,
				"machine_id", pid.MachineId)

			m, ok := getMachine(pid.MachineId)
			if ok && m.connected {
				okChan <- true

				m.monitorChan <- remoteMonitorTuple{From: c.self, To: pid}
				return
			}

			logger.Warn("remote machine is not registered, couldn't monitor pid",
				"monitored_gpid", pid.String(),
				"monitor_pid", c.self.Id,
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
			"monitored_gpid", pid.String(),
			"monitor_pid", c.self.Id)

		doSend(c.self, DownMessage{Who: pid}, nil)
		return &noopAbortable{}
	}
}
