package genserver

import (
	"errors"
	"github.com/Azer0s/quacktors"
	"reflect"
)

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
			Error: nil,
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
