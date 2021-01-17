package genserver

import (
	"errors"
	"github.com/Azer0s/quacktors"
	"reflect"
	"regexp"
)

var handleInfo = regexp.MustCompile("^Handle(.+)$")
var handleCast = regexp.MustCompile("^Handle(.+)Cast$")
var handleCall = regexp.MustCompile("^Handle(.+)Call$")

func checkHandlerTypes(mType reflect.Type) {
	if mType.NumIn() != 3 {
		panic("a GenServer handler has to have 2 parameters")
	}

	path := mType.In(1).Elem().String()

	if path != reflect.TypeOf(quacktors.Context{}).String() {
		panic("the first parameter of a GenServer handler has to be a *quacktors.Context")
	}
}

func checkCallHandler(mType reflect.Type) {
	if mType.NumOut() != 1 {
		panic("a GenServer call handler has to have one return value")
	}

	if mType.Out(0).String() != "quacktors.Message" {
		panic("a GenServer call handler must return a quacktors.Message")
	}
}

func setHandlerMethod(regex *regexp.Regexp, handlers *map[string]reflect.Value, mType reflect.Type, m reflect.Method) {
	handlerMessageName := mType.In(2).Name()
	messageType := regex.FindStringSubmatch(m.Name)[1]

	if handlerMessageName != messageType {
		panic("")
	}

	(*handlers)[messageType] = m.Func
}

//New creates a new GenServer. See the GenServer documentation
//for how to create a custom GenServer.
func New(server GenServer) quacktors.Actor {
	t := reflect.TypeOf(server)
	methods := t.NumMethod()

	infoHandlers := make(map[string]reflect.Value)
	castHandlers := make(map[string]reflect.Value)
	callHandlers := make(map[string]reflect.Value)

	var defaultInfoHandler reflect.Value
	defaultInfoHandlerMethod, defaultInfoHandlerSet := t.MethodByName("HandleInfo")
	if defaultInfoHandlerSet {
		checkHandlerTypes(defaultInfoHandlerMethod.Func.Type())
		defaultInfoHandler = defaultInfoHandlerMethod.Func
	}

	var defaultCastHandler reflect.Value
	defaultCastHandlerMethod, defaultCastHandlerSet := t.MethodByName("HandleCast")
	if defaultCastHandlerSet {
		checkHandlerTypes(defaultCastHandlerMethod.Func.Type())
		defaultCastHandler = defaultCastHandlerMethod.Func
	}

	var defaultCallHandler reflect.Value
	defaultCallHandlerMethod, defaultCallHandlerSet := t.MethodByName("HandleCall")
	if defaultCallHandlerSet {
		t := defaultCallHandlerMethod.Func.Type()

		checkHandlerTypes(t)
		checkCallHandler(t)

		defaultCallHandler = defaultCallHandlerMethod.Func
	}

	for i := 0; i < methods; i++ {
		m := t.Method(i)
		mType := m.Func.Type()

		if handleCall.MatchString(m.Name) {
			checkHandlerTypes(mType)
			checkCallHandler(mType)

			setHandlerMethod(handleCall, &callHandlers, mType, m)
			continue
		}

		if handleCast.MatchString(m.Name) {
			checkHandlerTypes(mType)
			setHandlerMethod(handleCast, &castHandlers, mType, m)
			continue
		}

		if handleInfo.MatchString(m.Name) {
			checkHandlerTypes(mType)
			setHandlerMethod(handleInfo, &infoHandlers, mType, m)
			continue
		}
	}

	return &genServerComponent{
		self:                  reflect.ValueOf(server),
		initFunction:          server.InitGenServer,
		infoHandlers:          infoHandlers,
		defaultInfoHandler:    defaultInfoHandler,
		defaultInfoHandlerSet: defaultInfoHandlerSet,
		callHandlers:          callHandlers,
		defaultCallHandler:    defaultCallHandler,
		defaultCallHandlerSet: defaultCallHandlerSet,
		castHandlers:          castHandlers,
		defaultCastHandler:    defaultCastHandler,
		defaultCastHandlerSet: defaultCastHandlerSet,
	}
}

//Call sends a message to the GenServer and blocks
//until the operation was completed by the GenServer
//and the GenServer returned a result. If there was
//an error, the GenServer went down or the PID was
//dead to begin with, Call returns an empty response
//message and a non-nil error. Otherwise the error
//is nil.
//This operation is blocking and should be used if
//you need to make sure a GenServer has processed a
//message.
func Call(context quacktors.Context, pid *quacktors.Pid, message quacktors.Message) (ResponseMessage, error) {
	returnChan := make(chan ResponseMessage)
	errChan := make(chan bool)

	p := quacktors.SpawnWithInit(func(ctx *quacktors.Context) {
		ctx.Monitor(pid)
	}, func(ctx *quacktors.Context, message quacktors.Message) {
		switch m := message.(type) {
		case ResponseMessage:
			returnChan <- m
		case quacktors.DownMessage:
			errChan <- true
		}

		ctx.Quit()
	})

	context.Send(pid, callMessage{
		sender:  p,
		message: message,
	})

	select {
	case res := <-returnChan:
		return res, nil
	case <-errChan:
		return ResponseMessage{}, errors.New("there was an error while calling to GenServer")
	}
}

//Cast sends a message to the GenServer and blocks
//until the GenServer has received the message and
//is about to start processing the it. If there was
//an error, the GenServer went down or the PID was
//dead to begin with, Cast returns a non-nil error.
//Otherwise the error is nil.
//This operation is blocking (if only for a very
//short time) and should be used if you need to make
//sure a GenServer has received a message but don't
//care whether the GenServer has failed or not.
func Cast(context quacktors.Context, pid *quacktors.Pid, message quacktors.Message) (ReceivedMessage, error) {
	returnChan := make(chan ReceivedMessage)
	errChan := make(chan bool)

	p := quacktors.SpawnWithInit(func(ctx *quacktors.Context) {
		ctx.Monitor(pid)
	}, func(ctx *quacktors.Context, message quacktors.Message) {
		switch m := message.(type) {
		case ReceivedMessage:
			returnChan <- m
		case quacktors.DownMessage:
			errChan <- true
		}

		ctx.Quit()
	})

	context.Send(pid, castMessage{
		sender:  p,
		message: message,
	})

	select {
	case res := <-returnChan:
		return res, nil
	case <-errChan:
		return ReceivedMessage{}, errors.New("there was an error while casting to GenServer")
	}
}
