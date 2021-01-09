package component

import (
	"github.com/Azer0s/quacktors"
	"github.com/Azer0s/quacktors/register"
)

//Relay returns a quacktors.Actor that forwards messages
//to a named actor.
func Relay(pidName string) *relayComponent {
	return &relayComponent{
		pidName: pidName,
	}
}

type relayComponent struct {
	pidName string
}

func (r *relayComponent) Init(ctx *quacktors.Context) {
}

func (r *relayComponent) Run(ctx *quacktors.Context, msg quacktors.Message) {
	register.UsePid(r.pidName, func(pid *quacktors.Pid) {
		ctx.Send(pid, msg)
	})
}
