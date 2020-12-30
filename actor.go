package quacktors

type Actor interface {
	Init(ctx *Context)
	Run(ctx *Context, message Message)
}

type StatelessActor struct {
	initFunction func(ctx *Context)
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
			//TODO: send to other machine
			return
		}

		to.messageChanMu.RLock()
		defer to.messageChanMu.RUnlock()

		if to.messageChan == nil {
			return
		}

		to.messageChan <- message
	}()
}

func cleanupActor(pid *Pid) {
	deletePid(pid.Id)

	//Terminate all scheduled events/send down message to monitor tasks
	for n, c := range pid.scheduled {
		c <- true //this is blocking
		close(c)
		delete(pid.scheduled, n)
	}

	//Delete monitorQuitChannels
	for n, c := range pid.monitorQuitChannels {
		close(c)
		delete(pid.monitorQuitChannels, n)
	}

	pid.monitorQuitChannels = nil

	pid.quitChanMu.Lock()
	close(pid.quitChan)
	pid.quitChan = nil
	pid.quitChanMu.Unlock()

	pid.messageChanMu.Lock()
	close(pid.messageChan)
	pid.messageChan = nil
	pid.messageChanMu.Unlock()

	pid.monitorChanMu.Lock()
	close(pid.monitorChan)
	pid.monitorChan = nil
	pid.monitorChanMu.Unlock()

	pid.demonitorChanMu.Lock()
	close(pid.demonitorChan)
	pid.demonitorChan = nil
	pid.demonitorChanMu.Unlock()
}

func setupMonitor(pid *Pid, monitor *Pid) {
	monitorChannel := make(chan bool)
	pid.scheduled[monitor.String()] = monitorChannel

	monitorQuitChannel := make(chan bool)
	pid.monitorQuitChannels[monitor.String()] = monitorQuitChannel

	go func() {
		select {
		case <- monitorQuitChannel:
			return
		case <- monitorChannel:
			doSend(monitor, &DownMessage{Who: pid})
		}
	}()
}

func removeMonitor(pid *Pid, monitor *Pid) {
	name := monitor.String()

	pid.monitorQuitChannels[name] <- true

	close(pid.monitorQuitChannels[name])
	close(pid.scheduled[name])

	delete(pid.monitorQuitChannels, name)
	delete(pid.scheduled, name)
}

func startActor(actor Actor) *Pid {
	quitChan := make(chan bool)             //channel to quit
	messageChan := make(chan Message, 2000) //channel for messages
	monitorChan := make(chan *Pid)          //channel to notify the actor of who wants to monitor it
	demonitorChan := make(chan *Pid)        //channel to notify the actor of who wants to unmonitor it

	scheduled := make(map[string]chan bool)
	monitorQuitChannels := make(map[string]chan bool)

	pid := createPid(quitChan, messageChan, monitorChan, demonitorChan, scheduled, monitorQuitChannels)
	registerPid(pid)

	ctx := &Context{self: pid}

	actor.Init(ctx)

	go func() {
		defer func() {
			//We don't want to forward a panic
			recover()

			cleanupActor(pid)
		}()

		for {
			select {
			case <-quitChan:
				return
			case message := <-messageChan:
				actor.Run(ctx, message)
			case monitor := <-monitorChan:
				setupMonitor(pid, monitor)
			case monitor := <-demonitorChan:
				removeMonitor(pid, monitor)
			}
		}
	}()

	return pid
}
