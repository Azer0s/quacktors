package quacktors

import (
	"encoding/gob"
	"errors"
	"github.com/rs/zerolog/log"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

func RegisterType(message Message) {
	t := reflect.ValueOf(message).Type().Kind()

	if t == reflect.Ptr {
		panic("RegisterType cannot be called with a pointer to a Message")
	}

	if message.Type() == "" {
		panic("message.Type() can not return an empty string")
	}

	gob.Register(message)

	log.Info().
		Str("type", message.Type()).
		Msg("registered type")
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
	log.Info().
		Str("system_name", name).
		Msg("initializing new system")

	s := &System{
		name:              name,
		handlers:          map[string]*Pid{},
		handlersMu:        &sync.RWMutex{},
		quitChan:          make(chan bool),
		heartbeatQuitChan: make(chan bool),
	}
	p, err := s.startServer()

	if err != nil {
		log.Warn().
			Str("system_name", s.name).
			Err(err).
			Msg("there was an error while starting the system server")
		return &System{}, err
	}

	conn, err := qpmdRegister(s, p)

	if err != nil {
		log.Warn().
			Str("system_name", s.name).
			Err(err).
			Msg("there was an error while registering system to qpmd")
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

	log.Info().
		Str("system_name", s[0]).
		Str("remote_address", s[1]).
		Msg("connecting to remote system")

	r, err := qpmdLookup(s[0], s[1])

	if err != nil {
		log.Warn().
			Str("system_name", s[0]).
			Str("remote_address", s[1]).
			Err(err).
			Msg("there was an error while looking up remote system")
		return &RemoteSystem{}, err
	}

	if r.MachineId == machineId {
		panic("can't connect to system, system is on the same quacktor instance")
	}

	if m, ok := getMachine(r.MachineId); ok {
		r.Machine = m
	} else {
		//start connections to remote machine

		log.Warn().
			Str("machine_id", r.MachineId).
			Msg("remote machine is not yet connected")

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
