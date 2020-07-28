package pid

// Pid is the interface type for the process ID
type Pid interface {
	Send(data interface{}, orderingComplete chan interface{})
	Monitor(by Pid)
	Up() bool
}
