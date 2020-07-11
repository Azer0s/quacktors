package actors

import "github.com/Azer0s/quacktors/pid"

type ActorDownMessage struct {
	Who pid.Pid
}
