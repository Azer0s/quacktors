package quacktors

import (
	"fmt"
)

//The Pid struct acts as a reference to an Actor.
//It is completely location transparent, meaning it doesn't
//matter if the Pid is actually on another system. To the
//developer it will look like just another Actor they can
//send messages to.
type Pid struct {
	MachineId     string
	Id            string
	quitChan      chan<- bool
	messageChan   chan<- localMessage
	monitorChan   chan<- *Pid
	demonitorChan chan<- *Pid
	//Stores channels to scheduled tasks (monitors, SendAfter, monitors the actor itself launches but doesn't consume)
	scheduled map[string]chan bool
	//Stores channels to tell a monitor taks to quit (when a pid is demonitored)
	monitorQuitChannels map[string]chan bool
}

func createPid(quitChan chan<- bool, messageChan chan<- localMessage, monitorChan chan<- *Pid, demonitorChan chan<- *Pid, scheduled map[string]chan bool, monitorQuitChannels map[string]chan bool) *Pid {
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

//Is compares two PIDs and returns true if their ID and MachineId are the same.
func (pid *Pid) Is(other *Pid) bool {
	return pid.Id == other.Id && pid.MachineId == other.MachineId
}

func (pid *Pid) cleanup() {
	logger.Debug("cleaning up pid",
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

	if len(pid.scheduled) != 0 {
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
	}

	if len(pid.monitorQuitChannels) != 0 {
		logger.Debug("deleting monitor abort channels",
			"pid_id", pid.Id)

		//Delete monitorQuitChannels
		for n, c := range pid.monitorQuitChannels {
			close(c)
			delete(pid.monitorQuitChannels, n)
		}
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
			doSend(monitor, DownMessage{Who: pid}, nil)
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

//Type returns the Message type of the PID.
//Since PIDs can be sent around without any message wrapper,
//it needs to implement the Message interface (which is why
//Type is needed).
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
