package test

import (
	"fmt"
	"github.com/Azer0s/quacktors"
	"github.com/Azer0s/quacktors/pid"
	"testing"
)

func TestActorCommunication(t *testing.T) {
	a := quacktors.Spawn(func() {
		bPid := quacktors.Receive().(pid.Pid)

		for {
			msg := quacktors.Receive()
			quacktors.Send(bPid, msg)
		}
	})

	b := quacktors.Spawn(func() {
		cPid := quacktors.Receive().(pid.Pid)

		for {
			msg := quacktors.Receive()
			quacktors.Send(cPid, msg)
		}
	})

	selfPid := quacktors.Self()

	c := quacktors.Spawn(func() {
		msg := quacktors.Receive()
		fmt.Println(msg)

		quacktors.Send(selfPid, nil)
	})

	quacktors.Send(a, b)
	quacktors.Send(b, c)

	quacktors.Send(a, "Hello")

	quacktors.Receive()
}
