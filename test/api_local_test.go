package test

import (
	"fmt"
	"github.com/Azer0s/quacktors"
	"github.com/Azer0s/quacktors/actors"
	"github.com/Azer0s/quacktors/pid"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestSelf(t *testing.T) {
	a := assert.New(t)

	self := quacktors.Self()
	self2 := quacktors.Self()

	a.Equal(self, self2)
}

func TestSpawn(t *testing.T) {
	a := assert.New(t)

	p := quacktors.Spawn(func() {
		for {
			fmt.Println("Hello")
		}
	})

	a.NotEqual(p, quacktors.Self())
}

func TestSend(t *testing.T) {
	a := assert.New(t)

	self := quacktors.Self()
	msg := "Hello, world!"

	quacktors.Spawn(func() {
		quacktors.Send(self, msg)
	})

	a.Equal(quacktors.Receive(), msg)
}

func TestSendReceive(t *testing.T) {
	wg := sync.WaitGroup{}
	a := assert.New(t)

	self := quacktors.Self()

	type message struct {
		sender  pid.Pid
		text string
	}

	wg.Add(1)
	quacktors.Spawn(func() {
		quacktors.Send(self, message{
			sender:  quacktors.Self(),
			text: "ping",
		})
		a.Equal(quacktors.Receive(), "pong")
		wg.Done()
	})

	msg := quacktors.Receive().(message)
	a.Equal("ping", msg.text)

	quacktors.Send(msg.sender, "pong")
	wg.Wait()
}

func TestMonitor(t *testing.T) {
	a := assert.New(t)

	p := quacktors.Spawn(func() {
		quacktors.Receive()
		//ignored
	})

	quacktors.Monitor(p)
	quacktors.Send(p, nil)

	msg := quacktors.Receive()

	if downMsg, ok := msg.(actors.ActorDownMessage); ok {
		a.Equal(downMsg.Who, p)
		a.False(p.Up())
	}
}