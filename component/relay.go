package component

import (
	"github.com/Azer0s/quacktors"
	"github.com/Azer0s/quacktors/register"
)

func Relay(pidName string) *RelayComponent {
	return &RelayComponent{
		pidName: pidName,
	}
}

type RelayComponent struct {
	pidName string
}

func (r *RelayComponent) Init(ctx *quacktors.Context) {
}

func (r *RelayComponent) Run(ctx *quacktors.Context, msg quacktors.Message) {
	register.UsePid(r.pidName, func(pid *quacktors.Pid) {
		ctx.Send(pid, msg)
	})
}
