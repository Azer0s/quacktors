package quacktors

import (
	"github.com/opentracing/opentracing-go"
	"sync"
)

//The Actor interface defines the methods a struct has to implement
//so it can be spawned by quacktors.
type Actor interface {
	//Init is called when an Actor is initialized. It is
	//guaranteed to be called before an Actor has been registered
	//or even started. Typically, Init is used to start monitors
	//to other actors or do some setup work. The caller
	//function provides a Context to the Init function.
	//Context can be used to interact with other actors
	//(e.g. send, monitor, etc) or modify the current Actor
	//(e.g. quit, defer actions, etc).
	Init(ctx *Context)

	//Run is called when an Actor receives a Message. The caller
	//function provides both a Context as well as the actual
	//Message to the Run function. Context can then be used to
	//interact with other actors (e.g. send, monitor, etc) or
	//modify the current Actor (e.g. quit, defer actions, etc).
	Run(ctx *Context, message Message)
}

//The StatelessActor struct is the Actor implementation that is
//used when using Spawn or SpawnWithInit. As the name implies,
//the StatelessActor doesn't have a state and just requires one
//anonymous function as the initializer (for Init) and another one
//as the run function (for Run) to work.
//ReceiveFunction can be nil, InitFunction has to be set.
type StatelessActor struct {
	InitFunction    func(ctx *Context)
	ReceiveFunction func(ctx *Context, message Message)
}

//Init initializes the StatelessActor by calling InitFunction if it
//is not nil. Init panics if ReceiveFunction is not set.
func (s *StatelessActor) Init(ctx *Context) {
	if s.InitFunction != nil {
		s.InitFunction(ctx)
	}

	if s.ReceiveFunction == nil {
		panic("ReceiveFunction of a StatelessActor cannot be nil")
	}
}

//Run forwards both the Message and the Context to the ReceiveFunction
//when the StatelessActor receives a message.
func (s *StatelessActor) Run(ctx *Context, message Message) {
	s.ReceiveFunction(ctx, message)
}

func doSend(to *Pid, message Message, spanContext opentracing.SpanContext) {
	returnChan := make(chan bool)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				//This happens if we write to the messageChan while the actor or remote connection is being closed
			}

			//As soon as we have put the message into the buffered messageChan, return
			//This is to preserve message ordering
			returnChan <- true
		}()

		if to.MachineId != machineId {
			//Pid is not on this machine

			m, ok := getMachine(to.MachineId)

			if ok && m.connected {
				m.messageChan <- remoteMessageTuple{
					To:          to,
					Message:     message,
					SpanContext: spanContext,
				}
			}

			return
		}

		//If the actor has already quit, do nothing
		if to.messageChan == nil {
			//Maybe the current pid instance is just empty but the pid actually does exist on our local machine
			//This can happen when you send the pid to a remote machine and receive it back
			p, ok := getByPidId(to.Id)

			if ok {
				p.messageChan <- localMessage{
					message:     message,
					spanContext: spanContext,
				}
			}

			return
		}

		to.messageChan <- localMessage{
			message:     message,
			spanContext: spanContext,
		}
	}()

	<-returnChan
}

func startActor(actor Actor) *Pid {
	quitChan := make(chan bool)                  //channel to quit
	messageChan := make(chan localMessage, 2000) //channel for messages
	monitorChan := make(chan *Pid)               //channel to notify the actor of who wants to monitor it
	demonitorChan := make(chan *Pid)             //channel to notify the actor of who wants to unmonitor it

	scheduled := make(map[string]chan bool)
	monitorQuitChannels := make(map[string]chan bool)

	pid := createPid(quitChan, messageChan, monitorChan, demonitorChan, scheduled, monitorQuitChannels)
	ctx := &Context{
		self:     pid,
		Logger:   contextLogger{pid: pid.Id},
		sendLock: &sync.Mutex{},
		deferred: make([]func(), 0),
	}

	actor.Init(ctx)

	logger.Info("starting actor",
		"pid", pid.String())

	go func() {
		defer func() {
			//We don't want to forward a panic
			if r := recover(); r != nil {
				if _, ok := r.(quitAction); ok {
					logger.Info("actor quit",
						"pid", pid.String())
				} else {
					//if we did pick up a panic, log it
					logger.Warn("actor quit due to panic",
						"pid", pid.String(),
						"panic", r)
				}
			}

			if len(ctx.deferred) != 0 {
				ctx.Logger.Debug("executing deferred actor actions")

				for _, action := range ctx.deferred {
					func() {
						defer func() {
							if r := recover(); r != nil {
								//action failed but we want to ignore that
							}
						}()
						action()
					}()
				}

				ctx.deferred = make([]func(), 0)
			}

			pid.cleanup()
		}()

		for {
			select {
			case <-quitChan:
				logger.Info("actor received quit event",
					"pid", pid.String())
				return
			case m := <-messageChan:
				switch m.message.(type) {
				case PoisonPill:
					logger.Info("actor received poison pill",
						"pid", pid.String())
					//Quit actor on PoisonPill message
					return
				default:
					ctx.span = nil

					func() {
						if m.spanContext != nil && ctx.traceName != "" {
							span := opentracing.GlobalTracer().StartSpan(ctx.traceName,
								opentracing.ChildOf(m.spanContext))
							span.SetTag("pid", pid.String())
							ctx.span = span

							defer span.Finish()
						}

						actor.Run(ctx, m.message)
					}()
				}
			case monitor := <-monitorChan:
				logger.Info("actor received monitor request",
					"pid", pid.String(),
					"monitor", monitor.String())
				pid.setupMonitor(monitor)
			case monitor := <-demonitorChan:
				logger.Info("actor received demonitor request",
					"pid", pid.String(),
					"monitor", monitor.String())
				pid.removeMonitor(monitor)
			}
		}
	}()

	return pid
}
