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
		defaultInfoHandler = defaultInfoHandlerMethod.Func
	}

	var defaultCastHandler reflect.Value
	defaultCastHandlerMethod, defaultCastHandlerSet := t.MethodByName("HandleCast")
	if defaultCastHandlerSet {
		defaultCastHandler = defaultCastHandlerMethod.Func
	}

	var defaultCallHandler reflect.Value
	defaultCallHandlerMethod, defaultCallHandlerSet := t.MethodByName("HandleCall")
	if defaultCallHandlerSet {
		defaultCallHandler = defaultCallHandlerMethod.Func
	}

	for i := 0; i < methods; i++ {
		m := t.Method(i)

		if handleCall.MatchString(m.Name) {
			messageType := handleCall.FindStringSubmatch(m.Name)[1]
			callHandlers[messageType] = m.Func
			continue
		}

		if handleCast.MatchString(m.Name) {
			messageType := handleCast.FindStringSubmatch(m.Name)[1]
			castHandlers[messageType] = m.Func
			continue
		}

		if handleInfo.MatchString(m.Name) {
			messageType := handleInfo.FindStringSubmatch(m.Name)[1]
			infoHandlers[messageType] = m.Func
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
func Call(pid *quacktors.Pid, message quacktors.Message) (ResponseMessage, error) {
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

	context := quacktors.RootContext()
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
func Cast(pid *quacktors.Pid, message quacktors.Message) (ReceivedMessage, error) {
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

	context := quacktors.RootContext()
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
