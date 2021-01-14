package component

import (
	"github.com/Azer0s/quacktors"
	"github.com/gofrs/uuid"
)

type actorPidTuple struct {
	name string
	pid  *quacktors.Pid
}

func DynamicSupervisor(strategy strategy, actors []quacktors.Actor) (supervisor quacktors.Actor, actorPids []*quacktors.Pid) {
	mapping := make([]actorPidTuple, 0)
	actorMap := make(map[string]quacktors.Actor)

	for _, actor := range actors {
		id, _ := uuid.NewV4()

		mapping = append(mapping, actorPidTuple{
			name: id.String(),
			pid:  nil,
		})
		actorMap[id.String()] = actor
	}

	supervisor = Supervisor(strategy, actorMap)
	actorPids = make([]*quacktors.Pid, 0)

	for _, tuple := range mapping {
		actorPids = append(actorPids, quacktors.SpawnStateful(Relay(tuple.name)))
	}

	return
}
