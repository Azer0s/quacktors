package genserver

import (
	"errors"
	"github.com/Azer0s/quacktors"
	"reflect"
)

//The GenServer interface defines the init method for a custom
//GenServer. The reason just the init method is defined, is that
//everything else is parsed out via reflection. This is also
//the reason why the GenServer has its own package (because
//it is way more complex than any other component).
//
//Communication types:
//
//There are 3 ways of how to communicate with a GenServer.
//The first one is via a Call. A Call is synchronous (meaning it
//will only return whenever the actual actor is done with whatever
//it was supposed to do). Calls are blocking but good if you need
//to know that a message was definitely processed.
//
//The second way of communicating with a GenServer is a cast.
//A Cast is partly synchronous (it will return after the GenServer
//has received the message and is about to start processing it).
//Casts are great when you only care about the actor receiving
//a message but not if the operation was successful.
//
//The third and final way is a normal send. This is completely
//asynchronous and acts like any other actor would (just that
//a GenServer offers some more framework sugar to make it easier
//to work with).
//
//Usage:
//
//The handlers links for a custom GenServer are described via
//the method names. The general format of a GenServer handler
//is:
//  Handle + MessageType + (Call | Cast |or nothing for a send handler)
//So to handle a GenericMessage Cast, the method name would look
//like so:
//  func (m myGenServer) HandleGenericMessageCast(ctx *Context, message GenericMessage)
//And a handler for a KillMessage Call would look like this:
//  func (m myGenServer) HandleKillMessageCall(ctx *Context, message KillMessage) Message
//A default handler for a DownMessage would look like this:
//  func (m myGenServer) HandleDownMessage(ctx *Context, message DownMessage)
//Note that the Call method returns a message, while the normal
//send handler and the Cast handler don't. This is because
//a Call is the only GenServer operation that can directly return
//something to the sender.
//
//You can optionally define "catch-all" handlers by leaving out
//the message type:
//  func (m myGenServer) HandleCast(ctx *Context, message GenericMessage)
//  func (m myGenServer) HandleCall(ctx *Context, message GenericMessage) Message
//  func (m myGenServer) HandleInfo(ctx *Context, message GenericMessage)
//(HandleInfo is a innuendo to the Elixir GenServer and is
//the "catch-all" for normal sends)
type GenServer interface {
	InitGenServer(ctx *quacktors.Context)
}

type genServerComponent struct {
	self                  reflect.Value
	initFunction          func(ctx *quacktors.Context)
	infoHandlers          map[string]reflect.Value
	defaultInfoHandler    reflect.Value
	defaultInfoHandlerSet bool
	callHandlers          map[string]reflect.Value
	defaultCallHandler    reflect.Value
	defaultCallHandlerSet bool
	castHandlers          map[string]reflect.Value
	defaultCastHandler    reflect.Value
	defaultCastHandlerSet bool
}

func (g *genServerComponent) Init(ctx *quacktors.Context) {

}

func (g *genServerComponent) doCall(ctx *quacktors.Context, m callMessage, handler reflect.Value) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				ctx.Send(m.sender, ResponseMessage{
					Message: nil,
					Error:   errors.New("GenServer went down during a call"),
				})
				panic(r)
			}
		}()

		res := handler.Call([]reflect.Value{g.self, reflect.ValueOf(ctx), reflect.ValueOf(m.message)})[0].Interface().(quacktors.Message)
		ctx.Send(m.sender, ResponseMessage{
			Message: res,
			Error:   nil,
		})
	}()
}

func (g *genServerComponent) Run(ctx *quacktors.Context, message quacktors.Message) {
	switch m := message.(type) {
	case callMessage:
		handler, ok := g.callHandlers[m.message.Type()]

		if ok {
			g.doCall(ctx, m, handler)
			return
		}

		if g.defaultCallHandlerSet {
			g.doCall(ctx, m, g.defaultCallHandler)
			return
		}

	case castMessage:
		handler, ok := g.castHandlers[m.message.Type()]

		if ok {
			ctx.Send(m.sender, ReceivedMessage{})
			handler.Call([]reflect.Value{g.self, reflect.ValueOf(ctx), reflect.ValueOf(m.message)})
			return
		}

		if g.defaultCastHandlerSet {
			ctx.Send(m.sender, ReceivedMessage{})
			g.defaultCastHandler.Call([]reflect.Value{g.self, reflect.ValueOf(ctx), reflect.ValueOf(m.message)})
			return
		}
	}

	handler, ok := g.infoHandlers[message.Type()]

	if ok {
		handler.Call([]reflect.Value{g.self, reflect.ValueOf(ctx), reflect.ValueOf(message)})
		return
	}

	if g.defaultInfoHandlerSet {
		g.defaultInfoHandler.Call([]reflect.Value{g.self, reflect.ValueOf(ctx), reflect.ValueOf(message)})
		return
	}
}
