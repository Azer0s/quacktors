package component

import (
	"github.com/Azer0s/quacktors"
	"github.com/Azer0s/quacktors/register"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRelayComponent(t *testing.T) {
	rootCtx := quacktors.RootContext()
	relayPid := quacktors.SpawnStateful(&RelayComponent{PidName: "foo"})

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

	quacktors.Wait()
}
