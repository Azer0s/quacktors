package pid

type Pid interface {
	Send(data interface{})
	Monitor(by Pid)
	Up() bool
}
