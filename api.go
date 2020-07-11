package quacktors

import (
	"github.com/Azer0s/quacktors/actors"
	"github.com/Azer0s/quacktors/node"
	"github.com/Azer0s/quacktors/pid"
	"github.com/Azer0s/quacktors/util"
	"sync"
)

// Self returns the PID of the caller goroutine/actor
func Self() pid.Pid {
	goid := util.GetGoid()
	p, err := actors.GetByGoid(goid)

	if err != nil {
		p = pid.NewPid()
		actors.StoreByGoid(goid, p)
	}

	return p
}

// Spawn spawns an actor by a function and returns the actors PID
func Spawn(action func()) pid.Pid {
	return actors.Spawn(action)
}

// Send sends data to a PID, this is a non-blocking call
func Send(pid pid.Pid, data interface{}) {
	go pid.Send(data)
}

// Receive receives data sent to the caller goroutine/actor, this is a blocking call
func Receive() interface{} {
	p := Self()
	return util.PidToLocalPid(p).Receive()
}

// Monitor monitors a PID, when the state of the monitored PID goes down, a message is sent to the monitoring actor
func Monitor(toMonitor pid.Pid) {
	p := Self()

	if !toMonitor.Up() {
		panic("Actor to monitor is down!")
	}

	toMonitor.Monitor(p)
}

func StartGateway(port int) {
	node.SetRemotePort(port)

	go func() {
		wg := sync.WaitGroup{}
		for {
			wg.Add(1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						//ignored
					}
				}()

				node.StartLink()
				wg.Done()
			}()
			wg.Wait()
		}
	}()
}
