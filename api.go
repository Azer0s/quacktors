package quacktors

import (
	"encoding/gob"
	"errors"
	"github.com/opentracing/opentracing-go"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

var initCalled = false

func callInitIfNotCalled() {
	if !initCalled {
		initQuacktorSystems()
	}

	initCalled = true
}

//RegisterType registers a Message to the type store so it can
//be sent to remote machines (which, of course, need a Message
//with the same Message.Type registered).
func RegisterType(message Message) {
	t := reflect.ValueOf(message).Type().Kind()

	if t == reflect.Ptr {
		panic("RegisterType cannot be called with a pointer to a Message")
	}

	if message.Type() == "" {
		panic("message.Type() can not return an empty string")
	}

	gob.RegisterName(message.Type(), message)

	logger.Info("registered type",
		"type", message.Type(),
	)
}

//RootContext returns a context that can be used outside an Actor.
//It is not associated with a real PID and therefore, one should
//not call anything with the RootContext that requires it to be
//able to receive or the like (e.g. no Context.Quit, Context.Monitor, etc.).
func RootContext() Context {
	callInitIfNotCalled()

	return Context{
		self:     &Pid{Id: "root", MachineId: machineId},
		Logger:   contextLogger{pid: "root"},
		sendLock: &sync.Mutex{},
		deferred: make([]func(), 0),
	}
}

//RootContextWithSpan is the same as RootContext but with a tracing
//span attached to it so it will do distributed tracing.
func RootContextWithSpan(span opentracing.Span) Context {
	callInitIfNotCalled()

	return Context{
		self:      &Pid{Id: "root", MachineId: machineId},
		Logger:    contextLogger{pid: "root"},
		sendLock:  &sync.Mutex{},
		deferred:  make([]func(), 0),
		span:      span,
		traceFork: opentracing.FollowsFrom,
	}
}

//VectorContext creates a context with a custom name and
//opentracing.Span. This allows for integrating applications
//with quacktors.
func VectorContext(name string, span opentracing.Span) Context {
	callInitIfNotCalled()

	return Context{
		self:      &Pid{Id: name, MachineId: machineId},
		Logger:    contextLogger{pid: name},
		sendLock:  &sync.Mutex{},
		deferred:  make([]func(), 0),
		span:      span,
		traceFork: opentracing.FollowsFrom,
	}
}

//Spawn spawns an Actor from an anonymous receive function and
//returns the *Pid of the Actor.
func Spawn(action func(ctx *Context, message Message)) *Pid {
	callInitIfNotCalled()

	return startActor(&StatelessActor{
		InitFunction:    func(ctx *Context) {},
		ReceiveFunction: action,
	})
}

//SpawnWithInit spawns an Actor from an anonymous receive function
//and an anonymous init function and returns the *Pid of the Actor.
func SpawnWithInit(init func(ctx *Context), action func(ctx *Context, message Message)) *Pid {
	callInitIfNotCalled()

	return startActor(&StatelessActor{
		InitFunction:    init,
		ReceiveFunction: action,
	})
}

//SpawnStateful spawns an Actor.
func SpawnStateful(actor Actor) *Pid {
	callInitIfNotCalled()

	return startActor(actor)
}

//NewSystem creates a new system server, connects to
//qpmd, starts the qpmd heartbeat for the new system and
//returns a *System and an error (nil if everything
//went fine).
func NewSystem(name string) (*System, error) {
	callInitIfNotCalled()

	logger.Info("initializing new system",
		"system_name", name)

	s := &System{
		name:              name,
		handlers:          make(map[string]*Pid),
		handlersMu:        &sync.RWMutex{},
		quitChan:          make(chan bool),
		heartbeatQuitChan: make(chan bool),
	}
	p, err := s.startServer()

	if err != nil {
		logger.Warn("there was an error while starting the system server",
			"system_name", s.name,
			"error", err)
		return &System{}, err
	}

	conn, err := qpmdRegister(s, p)

	if err != nil {
		logger.Warn("there was an error while registering system to qpmd",
			"system_name", s.name,
			"error", err)
		return &System{}, err
	}

	qpmdHeartbeat(conn, s)

	return s, nil
}

//Connect connects to a remote system and returns a
//*RemoteSystem and an error (nil if everything went
//fine). The connection string format should be
//"system@remote" where "system" is the name of the
//remote system and "remote" is either an IP or a
//domain name.
func Connect(name string) (*RemoteSystem, error) {
	callInitIfNotCalled()

	matched, err := regexp.MatchString("(\\w+)@(.+)", name)

	if !matched || err != nil {
		return &RemoteSystem{}, errors.New("invalid connection string format")
	}

	s := strings.SplitN(name, "@", 2)

	logger.Info("connecting to remote system",
		"system_name", s[0],
		"remote_address", s[1])

	r, err := qpmdLookup(s[0], s[1])

	if err != nil {
		logger.Warn("there was an error while looking up remote system",
			"system_name", s[0],
			"remote_address", s[1],
			"error", err)
		return &RemoteSystem{}, err
	}

	if r.MachineId == machineId {
		panic("can't connect to system, system is on the same quacktor instance")
	}

	if m, ok := getMachine(r.MachineId); ok {
		r.Machine = m
	} else {
		//start connections to remote machine

		logger.Warn("remote machine is not yet connected",
			"machine_id", r.MachineId)

		err := r.Machine.connect()

		if err != nil {
			return &RemoteSystem{}, err
		}

		registerMachine(r.Machine)
	}

	err = r.sayHello()

	if err != nil {
		return &RemoteSystem{}, err
	}

	return r, nil
}
