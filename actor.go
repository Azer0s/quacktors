package quacktors

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
	returnChan := make(chan bool)

	go func() {
		if to.MachineId != machineId {
			//Pid is not on this machine

			//Since we can't really guarantee message ordering to remote systems, this will have to do
			returnChan <- true

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

		//As soon as we have acquired the lock, return
		//This is to preserve message ordering
		returnChan <- true

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

	<-returnChan
}

func startActor(actor Actor) *Pid {
	quitChan := make(chan bool)             //channel to quit
	messageChan := make(chan Message, 2000) //channel for messages
	monitorChan := make(chan *Pid)          //channel to notify the actor of who wants to monitor it
	demonitorChan := make(chan *Pid)        //channel to notify the actor of who wants to unmonitor it

	scheduled := make(map[string]chan bool)
	monitorQuitChannels := make(map[string]chan bool)

	pid := createPid(quitChan, messageChan, monitorChan, demonitorChan, scheduled, monitorQuitChannels)
	ctx := &Context{
		self:   pid,
		Logger: contextLogger{pid: pid.Id},
	}

	actor.Init(ctx)

	logger.Info("starting actor",
		"pid", pid.String())

	go func() {
		defer func() {
			//We don't want to forward a panic
			recover()
			//TODO: if we did pick up a panic, log it
			pid.cleanup()
		}()

		for {
			select {
			case <-quitChan:
				logger.Info("actor received quit event",
					"pid", pid.String())
				return
			case message := <-messageChan:
				switch message.(type) {
				case PoisonPill:
					logger.Info("actor received poison pill",
						"pid", pid.String())
					//Quit actor on PoisonPill message
					return
				default:
					actor.Run(ctx, message)
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
