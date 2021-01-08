package component

import (
	"github.com/Azer0s/quacktors"
	"github.com/Azer0s/quacktors/register"
)

type RelayComponent struct {
	PidName string
}

func (r *RelayComponent) Init(ctx *quacktors.Context) {
}

func (r *RelayComponent) Run(ctx *quacktors.Context, msg quacktors.Message) {
	register.UsePid(r.PidName, func(pid *quacktors.Pid) {
		ctx.Send(pid, msg)
	})
}
