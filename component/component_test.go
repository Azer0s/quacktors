package component

import (
	"fmt"
	"github.com/Azer0s/quacktors"
	"github.com/Azer0s/quacktors/register"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestRelayComponent(t *testing.T) {
	rootCtx := quacktors.RootContext()
	relayPid := quacktors.SpawnStateful(Relay("foo"))

	p := quacktors.Spawn(func(ctx *quacktors.Context, message quacktors.Message) {
		msg := message.(quacktors.GenericMessage)

		self := ctx.Self()

		switch msg.Value.(string) {
		case "init_1":
			register.ChangePid("foo", func() *quacktors.Pid {
				return quacktors.Spawn(func(ctx *quacktors.Context, message quacktors.Message) {
					assert.Equal(t, "hello_1", message.(quacktors.GenericMessage).Value)
					ctx.Send(self, quacktors.GenericMessage{Value: "init_2"})
					ctx.Send(ctx.Self(), quacktors.PoisonPill{})
				})
			})
			ctx.Send(relayPid, quacktors.GenericMessage{Value: "hello_1"})

		case "init_2":
			register.ChangePid("foo", func() *quacktors.Pid {
				return quacktors.Spawn(func(ctx *quacktors.Context, message quacktors.Message) {
					assert.Equal(t, "hello_2", message.(quacktors.GenericMessage).Value)
					ctx.Send(self, quacktors.PoisonPill{})
					ctx.Send(ctx.Self(), quacktors.PoisonPill{})

					ctx.Kill(relayPid)
				})
			})
			ctx.Send(relayPid, quacktors.GenericMessage{Value: "hello_2"})
		}
	})

	rootCtx.Send(p, quacktors.GenericMessage{Value: "init_1"})

	quacktors.Run()
}

var count = 0

type testActor struct {
	id int
}

func (t *testActor) Init(ctx *quacktors.Context) {
	ctx.Logger.Info(fmt.Sprintf("started testActor %d", t.id))
	count++
}

func (t *testActor) Run(ctx *quacktors.Context, message quacktors.Message) {
	if _, ok := message.(quacktors.KillMessage); ok {
		ctx.Quit()
	}
}

func TestSupervisorOneForOne(t *testing.T) {
	count = 0

	rootCtx := quacktors.RootContext()

	supervisorPid := quacktors.SpawnStateful(Supervisor(ONE_FOR_ONE_STRATEGY, map[string]quacktors.Actor{
		"1": &testActor{id: 1},
		"2": &testActor{id: 2},
		"3": &testActor{id: 3},
		"4": &testActor{id: 4},
	}))

	register.UsePid("1", func(pid *quacktors.Pid) {
		rootCtx.Kill(pid)
	})

	register.UsePid("2", func(pid *quacktors.Pid) {
		rootCtx.Kill(pid)
	})

	register.UsePid("3", func(pid *quacktors.Pid) {
		rootCtx.Kill(pid)
	})

	register.UsePid("4", func(pid *quacktors.Pid) {
		rootCtx.Kill(pid)
	})

	<-time.After(1 * time.Second)

	assert.Equal(t, 8, count)

	rootCtx.Send(supervisorPid, quacktors.KillMessage{})

	quacktors.Run()
}

func TestSupervisorAllForOne(t *testing.T) {
	count = 0

	rootCtx := quacktors.RootContext()

	supervisorPid := quacktors.SpawnStateful(Supervisor(ALL_FOR_ONE_STRATEGY, map[string]quacktors.Actor{
		"1": &testActor{id: 1},
		"2": &testActor{id: 2},
		"3": &testActor{id: 3},
		"4": &testActor{id: 4},
	}))

	register.UsePid("1", func(pid *quacktors.Pid) {
		rootCtx.Kill(pid)
	})

	<-time.After(1 * time.Second)

	assert.Equal(t, 8, count)

	rootCtx.Send(supervisorPid, quacktors.KillMessage{})

	quacktors.Run()
}

func TestSupervisorFailAll(t *testing.T) {
	rootCtx := quacktors.RootContext()

	supervisorPid := quacktors.SpawnStateful(Supervisor(FAIL_ALL_STRATEGY, map[string]quacktors.Actor{
		"1": &testActor{id: 1},
		"2": &testActor{id: 2},
		"3": &testActor{id: 3},
		"4": &testActor{id: 4},
	}))

	p := quacktors.SpawnWithInit(func(ctx *quacktors.Context) {
		ctx.Monitor(supervisorPid)
	}, func(ctx *quacktors.Context, message quacktors.Message) {
		switch message.(type) {
		case quacktors.DownMessage:
			ctx.Quit()
		default:
			register.UsePid("1", func(pid *quacktors.Pid) {
				ctx.Kill(pid)
			})
		}
	})

	rootCtx.Send(p, quacktors.GenericMessage{})

	quacktors.Run()
}

func TestDynamicSupervisorOneForOne(t *testing.T) {
	count = 0

	rootCtx := quacktors.RootContext()

	dynamicSupervisor := DynamicSupervisor(ONE_FOR_ONE_STRATEGY, []quacktors.Actor{
		&testActor{id: 1},
		&testActor{id: 2},
		&testActor{id: 3},
		&testActor{id: 4},
	})

	supervisorPid := quacktors.SpawnStateful(dynamicSupervisor)
	pids := dynamicSupervisor.Pids()

	rootCtx.Send(pids[0], quacktors.KillMessage{})
	rootCtx.Send(pids[1], quacktors.KillMessage{})
	rootCtx.Send(pids[2], quacktors.KillMessage{})
	rootCtx.Send(pids[3], quacktors.KillMessage{})

	<-time.After(1 * time.Second)

	assert.Equal(t, 8, count)

	rootCtx.Send(supervisorPid, quacktors.KillMessage{})

	quacktors.Run()
}

func TestDynamicSupervisorAllForOne(t *testing.T) {
	count = 0

	rootCtx := quacktors.RootContext()

	dynamicSupervisor := DynamicSupervisor(ALL_FOR_ONE_STRATEGY, []quacktors.Actor{
		&testActor{id: 1},
		&testActor{id: 2},
		&testActor{id: 3},
		&testActor{id: 4},
	})

	supervisorPid := quacktors.SpawnStateful(dynamicSupervisor)
	pids := dynamicSupervisor.Pids()

	rootCtx.Send(pids[0], quacktors.KillMessage{})

	<-time.After(1 * time.Second)

	assert.Equal(t, 8, count)

	rootCtx.Send(supervisorPid, quacktors.KillMessage{})

	quacktors.Run()
}

func TestDynamicSupervisorFailAll(t *testing.T) {
	rootCtx := quacktors.RootContext()

	dynamicSupervisor := DynamicSupervisor(FAIL_ALL_STRATEGY, []quacktors.Actor{
		&testActor{id: 1},
		&testActor{id: 2},
		&testActor{id: 3},
		&testActor{id: 4},
	})

	supervisorPid := quacktors.SpawnStateful(dynamicSupervisor)
	pids := dynamicSupervisor.Pids()

	p := quacktors.SpawnWithInit(func(ctx *quacktors.Context) {
		ctx.Monitor(supervisorPid)
	}, func(ctx *quacktors.Context, message quacktors.Message) {
		switch message.(type) {
		case quacktors.DownMessage:
			ctx.Quit()
		default:
			ctx.Send(pids[0], quacktors.KillMessage{})
		}
	})

	rootCtx.Send(p, quacktors.GenericMessage{})

	quacktors.Run()
}

func TestLink(t *testing.T) {
	wg := &sync.WaitGroup{}

	p1 := quacktors.SpawnWithInit(func(ctx *quacktors.Context) {
		wg.Add(1)
		ctx.Defer(func() {
			wg.Done()
		})
	}, func(ctx *quacktors.Context, message quacktors.Message) {

	})

	p2 := quacktors.SpawnWithInit(func(ctx *quacktors.Context) {
		wg.Add(1)
		ctx.Defer(func() {
			wg.Done()
		})
	}, func(ctx *quacktors.Context, message quacktors.Message) {

	})

	quacktors.SpawnStateful(Link(p1, p2))

	context := quacktors.RootContext()
	context.Kill(p1)

	wg.Wait()
	quacktors.Run()
}

func TestLoadBalancer(t *testing.T) {
	count = 0
	usage := 1

	lb := LoadBalancer(10, &testActor{id: 1}, func() uint16 {
		return uint16(usage)
	})

	//load balancer should spin up 1 testActor
	lbPid := quacktors.SpawnStateful(lb)
	context := quacktors.RootContext()

	usage = 11

	//load balancer should spin up a second testActor
	context.Send(lbPid, quacktors.GenericMessage{})

	//load balancer should repair itself
	context.Send(lbPid, quacktors.KillMessage{})
	<-time.After(1 * time.Second)
	assert.Equal(t, 3, count)

	usage = 21

	//load balancer should spin up a third testActor
	context.Send(lbPid, quacktors.GenericMessage{})
	<-time.After(1 * time.Second)
	assert.Equal(t, 4, count)

	context.Send(lbPid, quacktors.PoisonPill{})
	quacktors.Run()
}
