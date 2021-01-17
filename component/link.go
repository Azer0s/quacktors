package component

import "github.com/Azer0s/quacktors"

//Link links two PIDs together, so when
//one goes down, the other one does too.
func Link(from *quacktors.Pid, to *quacktors.Pid) quacktors.Actor {
	return &linkComponent{
		from: from,
		to:   to,
	}
}

type linkComponent struct {
	from          *quacktors.Pid
	to            *quacktors.Pid
	fromAbortable quacktors.Abortable
	toAbortable   quacktors.Abortable
}

func (l *linkComponent) Init(ctx *quacktors.Context) {
	l.fromAbortable = ctx.Monitor(l.from)
	l.toAbortable = ctx.Monitor(l.to)
}

func (l *linkComponent) Run(ctx *quacktors.Context, message quacktors.Message) {
	d := message.(quacktors.DownMessage)

	if d.Who.Is(l.from) {
		l.toAbortable.Abort()
		ctx.Kill(l.to)
		ctx.Quit()
		return
	}

	l.fromAbortable.Abort()
	ctx.Kill(l.from)
	ctx.Quit()
}
