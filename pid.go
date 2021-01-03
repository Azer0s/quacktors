package quacktors

import (
	"fmt"
)

type Pid struct {
	MachineId     string
	Id            string
	quitChan      chan<- bool
	messageChan   chan<- Message
	monitorChan   chan<- *Pid
	demonitorChan chan<- *Pid
	//Stores channels to scheduled tasks (monitors, SendAfter, monitors the actor itself launches but doesn't consume)
	scheduled map[string]chan bool
	//Stores channels to tell a monitor taks to quit (when a pid is demonitored)
	monitorQuitChannels map[string]chan bool
}

func createPid(quitChan chan<- bool, messageChan chan<- Message, monitorChan chan<- *Pid, demonitorChan chan<- *Pid, scheduled map[string]chan bool, monitorQuitChannels map[string]chan bool) *Pid {
	pid := &Pid{
		MachineId:           machineId,
		Id:                  "",
		quitChan:            quitChan,
		messageChan:         messageChan,
		monitorChan:         monitorChan,
		demonitorChan:       demonitorChan,
		scheduled:           scheduled,
		monitorQuitChannels: monitorQuitChannels,
	}

	registerPid(pid)

	return pid
}

func (pid *Pid) cleanup() {
	logger.Info("cleaning up pid",
		"pid_id", pid.Id)

	deletePid(pid.Id)

	close(pid.quitChan)
	pid.quitChan = nil

	close(pid.messageChan)
	pid.messageChan = nil

	close(pid.monitorChan)
	pid.monitorChan = nil

	close(pid.demonitorChan)
	pid.demonitorChan = nil

	//Terminate all scheduled events/send down message to monitor tasks
	logger.Debug("sending out scheduled events after pid cleanup",
		"pid_id", pid.Id)

	for n, ch := range pid.scheduled {
		//what if someone aborts the monitor while we attempt to write to it?
		//this can never happen because all monitor and demonitor requests go
		//through the actor which is currently being closed

		ch <- true //this is blocking
		close(ch)
		delete(pid.scheduled, n)
	}

	logger.Debug("deleting monitor abort channels",
		"pid_id", pid.Id)

	//Delete monitorQuitChannels
	for n, c := range pid.monitorQuitChannels {
		close(c)
		delete(pid.monitorQuitChannels, n)
	}
	pid.monitorQuitChannels = nil
}

func (pid *Pid) setupMonitor(monitor *Pid) {
	//there used to be a mutex here but since all monitor and demonitor
	//requests go through one actor, we can't run into a concurrent rw

	name := monitor.String()

	monitorChannel := make(chan bool)
	pid.scheduled[name] = monitorChannel

	monitorQuitChannel := make(chan bool)
	pid.monitorQuitChannels[name] = monitorQuitChannel

	go func() {
		select {
		case <-monitorQuitChannel:
			return
		case <-monitorChannel:
			doSend(monitor, DownMessage{Who: pid})
		}
	}()
}

func (pid *Pid) removeMonitor(monitor *Pid) {
	name := monitor.String()

	pid.monitorQuitChannels[name] <- true

	close(pid.monitorQuitChannels[name])
	close(pid.scheduled[name])

	delete(pid.monitorQuitChannels, name)
	delete(pid.scheduled, name)

	logger.Info("monitor removed successfully",
		"monitored_pid", pid.String(),
		"monitor_pid", monitor.String())
}

func (pid *Pid) String() string {
	return fmt.Sprintf("%s@%s", pid.Id, pid.MachineId)
}

func (pid Pid) Type() string {
	return "pid"
}

func (pid *Pid) die() {
	defer func() {
		if r := recover(); r != nil {
			//This happens if we write to the quitChan while the actor is being closed
		}
	}()

	logger.Debug("sending quit command to actor",
		"pid", pid.String())

	if pid.quitChan == nil {
		return
	}

	pid.quitChan <- true
}
