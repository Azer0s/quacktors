package quacktors

import (
	"github.com/Azer0s/Quacktors/quacktors/actors"
	"github.com/Azer0s/Quacktors/quacktors/pid"
	"github.com/Azer0s/Quacktors/quacktors/util"
)

func Self() pid.Pid {
	goid := util.GetGoid()
	p, err := actors.GetByGoid(goid)

	if err != nil {
		p = pid.NewPid()
		actors.StoreByGoid(goid, p)
	}

	return p
}

func Spawn(action func()) pid.Pid {
	return actors.Spawn(action)
}

func Send(pid pid.Pid, data interface{}) {
	pid.Send(data)
}

func Receive() interface{} {
	p := Self()
	return util.PidToLocalPid(p).Receive()
}

func Monitor(toMonitor pid.Pid) {
	p := Self()

	if !toMonitor.Up() {
		panic("Actor to monitor is down!")
	}

	toMonitor.Monitor(p)
}