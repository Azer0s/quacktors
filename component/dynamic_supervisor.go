package component

import (
	"github.com/Azer0s/quacktors"
	"github.com/gofrs/uuid"
)

type actorPidTuple struct {
	name string
	pid  *quacktors.Pid
}

type dynamicSupervisorComponent struct {
	supervisor    quacktors.Actor
	supervisorPid *quacktors.Pid
	actorPids     []*quacktors.Pid
	mapping       []actorPidTuple
}

func (d *dynamicSupervisorComponent) Init(ctx *quacktors.Context) {
	d.actorPids = make([]*quacktors.Pid, 0)

	for _, tuple := range d.mapping {
		relayPid := quacktors.SpawnStateful(Relay(tuple.name))
		d.actorPids = append(d.actorPids, relayPid)

		ctx.Defer(func() {
			ctx.Kill(relayPid)
		})
	}

	d.supervisorPid = quacktors.SpawnStateful(d.supervisor)

	ctx.Defer(func() {
		ctx.Kill(d.supervisorPid)
	})

	ctx.Monitor(d.supervisorPid)
}

func (d *dynamicSupervisorComponent) Run(ctx *quacktors.Context, msg quacktors.Message) {
	switch m := msg.(type) {
	case quacktors.DownMessage:
		if m.Who.Is(d.supervisorPid) {
			ctx.Kill(ctx.Self())
		}
	case quacktors.KillMessage:
		ctx.Send(ctx.Self(), quacktors.PoisonPill{})
	}
}

func (d *dynamicSupervisorComponent) Pids() []*quacktors.Pid {
	return d.actorPids
}

//DynamicSupervisor returns a dynamic supervisor component.
//It is functionally the same as the Supervisor with the key
//difference being that child actors of the dynamic supervisor
//don't use named actors but can be addressed via PIDs
//(internally it still uses named actors but automatically creates
//and manages relays for each child).
func DynamicSupervisor(strategy strategy, actors []quacktors.Actor) *dynamicSupervisorComponent {
	d := dynamicSupervisorComponent{}

	d.mapping = make([]actorPidTuple, 0)
	actorMap := make(map[string]quacktors.Actor)

	for _, actor := range actors {
		id, _ := uuid.NewV4()

		d.mapping = append(d.mapping, actorPidTuple{
			name: id.String(),
			pid:  nil,
		})
		actorMap[id.String()] = actor
	}

	d.supervisor = Supervisor(strategy, actorMap)

	return &d
}
