package quacktors

import (
	"encoding/gob"
	"errors"
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

var rootContext = Context{
	self:     &Pid{Id: "root"},
	Logger:   contextLogger{pid: "root"},
	sendLock: &sync.Mutex{},
}

func RootContext() Context {
	callInitIfNotCalled()

	return rootContext
}

func Spawn(action func(ctx *Context, message Message)) *Pid {
	callInitIfNotCalled()

	return startActor(&StatelessActor{
		initFunction:    func(ctx *Context) {},
		receiveFunction: action,
	})
}

func SpawnWithInit(init func(ctx *Context), action func(ctx *Context, message Message)) *Pid {
	callInitIfNotCalled()

	return startActor(&StatelessActor{
		initFunction:    init,
		receiveFunction: action,
	})
}

func SpawnStateful(actor Actor) *Pid {
	callInitIfNotCalled()

	return startActor(actor)
}

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
