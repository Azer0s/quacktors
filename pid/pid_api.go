package pid

// NewPid creates a new local PID
func NewPid() Pid {
	return &LocalPid{incoming: make(chan interface{}, 1024 /*TODO: Refactor into constant*/), up: true}
}
