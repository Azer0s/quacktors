package quacktors

import (
	"errors"
	"go.uber.org/zap"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

func RegisterType(message Message) {
	t := reflect.ValueOf(message).Type().Kind()

	if t != reflect.Ptr {
		panic("RegisterType has to be called with a pointer to a Message")
	}

	if message.Type() == "" {
		panic("message.Type() can not return an empty string")
	}

	storeType(message)

	logger.Info("registered type", zap.String("type", message.Type()))
}

func RootContext() Context {
	return Context{}
}

func Spawn(action func(ctx *Context, message Message)) *Pid {
	return startActor(&StatelessActor{
		initFunction:    func(ctx *Context) {},
		receiveFunction: action,
	})
}

func SpawnWithInit(init func(ctx *Context), action func(ctx *Context, message Message)) *Pid {
	return startActor(&StatelessActor{
		initFunction:    init,
		receiveFunction: action,
	})
}

func SpawnStateful(actor Actor) *Pid {
	return startActor(actor)
}

func NewSystem(name string) (*System, error) {
	logger.Info("initializing new system", zap.String("system_name", name))

	s := &System{
		name:              name,
		handlers:          map[string]*Pid{},
		handlersMu:        &sync.RWMutex{},
		quitChan:          make(chan bool),
		heartbeatQuitChan: make(chan bool),
	}
	p, err := s.startServer()

	if err != nil {
		logger.Warn("there was an error while starting the system server",
			zap.String("system_name", s.name),
			zap.Error(err),
		)
		return &System{}, err
	}

	conn, err := qpmdRegister(s, p)

	if err != nil {
		logger.Warn("there was an error while registering system to qpmd",
			zap.String("system_name", s.name),
			zap.Error(err),
		)
		return &System{}, err
	}

	qpmdHeartbeat(conn, s)

	return s, nil
}

func Connect(name string) (*RemoteSystem, error) {
	matched, err := regexp.MatchString("(\\w+)@(.+)", name)

	if !matched || err != nil {
		return &RemoteSystem{}, errors.New("invalid connection string format")
	}

	s := strings.SplitN(name, "@", 2)

	logger.Info("connecting to remote system",
		zap.String("system_name", s[0]),
		zap.String("remote_address", s[1]),
	)

	r, err := qpmdLookup(s[0], s[1])

	if err != nil {
		logger.Warn("there was an error while looking up remote system",
			zap.String("system_name", s[0]),
			zap.String("remote_address", s[1]),
			zap.Error(err),
		)
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
			zap.String("machine_id", m.MachineId),
		)

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
