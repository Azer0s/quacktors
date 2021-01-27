package quacktors

import (
	"github.com/Azer0s/quacktors/mailbox"
	"github.com/Azer0s/quacktors/metrics"
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
				metrics.RecordUnhandled(to.Id)
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

				metrics.RecordSendRemote(to.Id)
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
				metrics.RecordSendLocal(p.Id)
			}

			return
		}

		to.messageChan <- localMessage{
			message:     message,
			spanContext: spanContext,
		}
		metrics.RecordSendLocal(to.Id)
	}()

	<-returnChan
}

func startActor(actor Actor) *Pid {
	quitChan := make(chan bool)      //channel to quit
	mb := mailbox.New()              //message mailbox
	monitorChan := make(chan *Pid)   //channel to notify the actor of who wants to monitor it
	demonitorChan := make(chan *Pid) //channel to notify the actor of who wants to unmonitor it

	scheduled := make(map[string]chan bool)
	monitorQuitChannels := make(map[string]chan bool)

	pid := createPid(quitChan, mb.In(), monitorChan, demonitorChan, scheduled, monitorQuitChannels)
	ctx := &Context{
		self:      pid,
		Logger:    contextLogger{pid: pid.Id},
		sendLock:  &sync.Mutex{},
		deferred:  make([]func(), 0),
		traceFork: opentracing.FollowsFrom,
	}

	//Initialize the actor
	actor.Init(ctx)

	//If the init was successful, record the spawn
	metrics.RecordSpawn(pid.Id)

	logger.Info("starting actor",
		"pid", pid.Id)

	messageChan := mb.Out()

	go func() {
		defer func() {
			//We don't want to forward a panic
			if r := recover(); r != nil {
				if _, ok := r.(quitAction); ok {
					logger.Info("actor quit",
						"pid", pid.Id)
				} else {
					//if we did pick up a panic, log it
					logger.Warn("actor quit due to panic",
						"pid", pid.Id,
						"panic", r)
				}
			}

			//We don't really care how the actor died, we just wanna know that it did
			metrics.RecordDie(pid.Id)

			unreadMessages := mb.Len()
			if unreadMessages != 0 {
				//If we still have pending messages in the channel, these are marked as dropped
				metrics.RecordDrop(pid.Id, mb.Len())
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
					"pid", pid.Id)
				return
			case mi := <-messageChan:
				metrics.RecordReceive(pid.Id)

				m := mi.(localMessage)

				switch m.message.(type) {
				case PoisonPill:
					logger.Info("actor received poison pill",
						"pid", pid.Id)
					//Quit actor on PoisonPill message
					return
				default:
					ctx.span = nil

					func() {
						if m.spanContext != nil && ctx.traceName != "" {
							span := opentracing.GlobalTracer().StartSpan(ctx.traceName,
								ctx.traceFork(m.spanContext))
							span.SetTag("pid", pid.Id)
							span.SetTag("machine_id", pid.MachineId)
							ctx.span = span

							defer span.Finish()
						}

						actor.Run(ctx, m.message)
					}()

					//Clean after run so the span won't be sent in any defers if the actor goes down right after
					ctx.span = nil
				}
			case monitor := <-monitorChan:
				logger.Info("actor received monitor request",
					"pid", pid.Id,
					"monitor_gpid", monitor.String())
				pid.setupMonitor(monitor)
			case monitor := <-demonitorChan:
				logger.Info("actor received demonitor request",
					"pid", pid.Id,
					"monitor_gpid", monitor.String())
				pid.removeMonitor(monitor)
			}
		}
	}()

	return pid
}
