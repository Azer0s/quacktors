package pid

import (
	"github.com/google/uuid"
)

// NewPid creates a new local PID
func NewPid() (Pid, string) {
	id := uuid.New().String()
	return &LocalPid{incoming: make(chan interface{}, 1024 /*TODO: Refactor into constant*/), up: true, id: id}, id
}
