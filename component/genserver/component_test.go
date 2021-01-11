package genserver

import (
	"fmt"
	"github.com/Azer0s/quacktors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testGenServer struct {
}

func (t testGenServer) InitGenServer(ctx *quacktors.Context) {

}

func (t testGenServer) HandleGenericMessage(ctx *quacktors.Context, message quacktors.GenericMessage) {
	assert.True(message.Value.(*testing.T), true)
	ctx.Quit()
}

func (t testGenServer) HandleEmptyMessageCast(ctx *quacktors.Context, message quacktors.EmptyMessage) {
	fmt.Println("empty")
}

func (t testGenServer) HandleGenericMessageCall(ctx *quacktors.Context, message quacktors.GenericMessage) quacktors.Message {
	fmt.Println(message.Value)

	return quacktors.GenericMessage{Value: message.Value.(string) + " back!"}
}

func TestGenServerCast(t *testing.T) {
	genServerPid := quacktors.SpawnStateful(New(testGenServer{}))

	r, _ := Cast(genServerPid, quacktors.EmptyMessage{})

	assert.Equal(t, "ReceivedMessage", r.Type())

	context := quacktors.RootContext()
	context.Kill(genServerPid)

	quacktors.Wait()
}

func TestGenServerCall(t *testing.T) {
	genServerPid := quacktors.SpawnStateful(New(testGenServer{}))

	r, _ := Call(genServerPid, quacktors.GenericMessage{Value: "Hi"})

	assert.Equal(t, "Hi back!", r.Message.(quacktors.GenericMessage).Value)

	context := quacktors.RootContext()
	context.Kill(genServerPid)

	quacktors.Wait()
}

func TestGenServerInfo(t *testing.T) {
	genServerPid := quacktors.SpawnStateful(New(testGenServer{}))

	context := quacktors.RootContext()
	context.Send(genServerPid, quacktors.GenericMessage{Value: t})

	quacktors.Wait()
}

func TestDeadGenServerCast(t *testing.T) {
	genServerPid := quacktors.SpawnStateful(New(testGenServer{}))
	context := quacktors.RootContext()
	context.Send(genServerPid, quacktors.PoisonPill{})

	_, err := Cast(genServerPid, quacktors.EmptyMessage{})

	assert.Error(t, err)

	quacktors.Wait()
}

func TestDeadGenServerCall(t *testing.T) {
	genServerPid := quacktors.SpawnStateful(New(testGenServer{}))
	context := quacktors.RootContext()
	context.Send(genServerPid, quacktors.PoisonPill{})

	_, err := Call(genServerPid, quacktors.GenericMessage{Value: "Hi"})

	assert.Error(t, err)

	quacktors.Wait()
}
