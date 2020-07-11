package actors

import "github.com/Azer0s/Quacktors/quacktors/pid"

type ActorDownMessage struct {
	Who pid.Pid
}
