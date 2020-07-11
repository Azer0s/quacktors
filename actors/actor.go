package actors

import (
	"github.com/Azer0s/quacktors/messages"
	"github.com/Azer0s/quacktors/pid"
	"github.com/Azer0s/quacktors/util"
)

// Spawn spawns an actor by a function and returns its PID
func Spawn(action func()) pid.Pid {
	p := pid.NewPid()
	Start(action, p)
	return p
}

// Start Starts the control function of an actor (does monitoring, etc)
func Start(action func(), pid pid.Pid) {
	go func() {
		goid := util.GetGoid()
		StoreByGoid(goid, pid)

		defer func() {
			DeleteByGoid(goid)

			localPid := util.PidToLocalPid(pid)
			localPid.Down()

			downMessage := messages.ActorDownMessage{Who: pid}
			for _, monitor := range localPid.Monitors() {
				go monitor.Send(downMessage)
			}
		}()

		action()
	}()
}
