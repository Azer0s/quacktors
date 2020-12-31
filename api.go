package quacktors

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

func RegisterType(message Message) {
	t := reflect.ValueOf(message).Type().Kind()

	if t != reflect.Ptr {
		panic("RegisterType has to be called with a pointer to a Message!")
	}
	storeType(message)
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
	s := &System{
		name:              name,
		handlers:          map[string]*Pid{},
		handlersMu:        &sync.RWMutex{},
		quitChan:          make(chan bool),
		heartbeatQuitChan: make(chan bool),
	}
	p, err := s.startServer()

	if err != nil {
		return &System{}, err
	}

	conn, err := qpmdRegister(s, p)

	if err != nil {
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

	r, err := qpmdLookup(s[0], s[1])

	if err != nil {
		return &RemoteSystem{}, err
	}

	err = r.sayHello()

	if err != nil {
		return &RemoteSystem{}, err
	}

	//TODO: start connections to remote machine if they don't exist yet

	return r, nil
}
