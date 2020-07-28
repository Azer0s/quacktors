package messages

import (
	"github.com/Azer0s/quacktors/pid"
)

// ActorDownMessage will be sent out to monitoring actors when a monitored actor goes down
type ActorDownMessage struct {
	Who pid.Pid
}
