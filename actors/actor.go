package actors

import (
	"github.com/Azer0s/quacktors/pid"
	"github.com/Azer0s/quacktors/util"
)

func Spawn(action func()) pid.Pid {
	p := pid.NewPid()
	Start(action, p)
	return p
}

func Start(action func(), pid pid.Pid) {
	go func() {
		goid := util.GetGoid()
		StoreByGoid(goid, pid)

		defer func() {
			DeleteByGoid(goid)

			localPid := util.PidToLocalPid(pid)
			localPid.Down()

			downMessage := ActorDownMessage{Who: pid}
			for _, monitor := range localPid.Monitors() {
				go monitor.Send(downMessage)
			}
		}()

		action()
	}()
}
