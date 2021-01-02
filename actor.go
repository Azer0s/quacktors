package quacktors

import (
	"github.com/rs/zerolog/log"
)

type Actor interface {
	Init(ctx *Context)
	Run(ctx *Context, message Message)
}

type StatelessActor struct {
	initFunction    func(ctx *Context)
	receiveFunction func(ctx *Context, message Message)
}

func (s *StatelessActor) Init(ctx *Context) {
	s.initFunction(ctx)
}

func (s *StatelessActor) Run(ctx *Context, message Message) {
	s.receiveFunction(ctx, message)
}

func doSend(to *Pid, message Message) {
	go func() {
		if to.MachineId != machineId {
			//Pid is not on this machine

			m, ok := getMachine(to.MachineId)

			if ok {
				m.messageChan <- remoteMessageTuple{
					To:      to,
					Message: message,
				}
			}

			return
		}

		//Lock the channel so we don't run into problems if we're in the middle of an actor quit
		to.messageChanMu.RLock()
		defer to.messageChanMu.RUnlock()

		//If the actor has already quit, do nothing
		if to.messageChan == nil {
			//Maybe the current pid instance is just empty but the pid actually does exist on our local machine
			//This can happen when you send the pid to a remote machine and receive it back
			p, ok := getByPidId(to.Id)

			if ok {
				p.messageChan <- message
			}

			return
		}

		to.messageChan <- message
	}()
}

func startActor(actor Actor) *Pid {
	quitChan := make(chan bool)             //channel to quit
	messageChan := make(chan Message, 2000) //channel for messages
	monitorChan := make(chan *Pid)          //channel to notify the actor of who wants to monitor it
	demonitorChan := make(chan *Pid)        //channel to notify the actor of who wants to unmonitor it

	scheduled := make(map[string]chan bool)
	monitorQuitChannels := make(map[string]chan bool)

	pid := createPid(quitChan, messageChan, monitorChan, demonitorChan, scheduled, monitorQuitChannels)
	ctx := &Context{self: pid}

	actor.Init(ctx)

	log.Info().
		Str("pid", pid.String()).
		Msg("starting actor")

	go func() {
		defer func() {
			//We don't want to forward a panic
			recover()
			pid.cleanup()
		}()

		for {
			select {
			case <-quitChan:
				log.Info().
					Str("pid", pid.String()).
					Msg("actor received quit event")
				return
			case message := <-messageChan:
				switch message.(type) {
				case *PoisonPill:
					log.Info().
						Str("pid", pid.String()).
						Msg("actor received poison pill")
					//Quit actor on PoisonPill message
					return
				default:
					actor.Run(ctx, message)
				}
			case monitor := <-monitorChan:
				log.Info().
					Str("pid", pid.String()).
					Str("monitor", monitor.String()).
					Msg("actor received monitor request")
				pid.setupMonitor(monitor)
			case monitor := <-demonitorChan:
				log.Info().
					Str("pid", pid.String()).
					Str("monitor", monitor.String()).
					Msg("actor received demonitor request")
				pid.removeMonitor(monitor)
			}
		}
	}()

	return pid
}
