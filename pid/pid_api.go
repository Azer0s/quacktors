package pid

func NewPid() Pid {
	return &LocalPid{incoming: make(chan interface{}, 1024 /*TODO: Refactor into constant*/), up: true}
}
