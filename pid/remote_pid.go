package pid

// RemotePid a PID that points to an actor on a remote system
type RemotePid struct {
}

// Send sends data to an actor on a remote system
func (p RemotePid) Send(data interface{}, orderingComplete chan interface{}) {
    // setup connection
    // p.connMu.Lock()
    // defer p.connMu.Unlock()
    // orderingComplete <- nil
    // send data
}

// Up returns true if an actor on a remote system is up
func (p RemotePid) Up() bool {
	return false
}

// Monitor monitors an actor on a remote system
func (p RemotePid) Monitor(by Pid) {

}
