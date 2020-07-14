package pid

import "sync"

// LocalPid is a PID that points to an actor on the local system
type LocalPid struct {
	incoming   chan interface{}
	monitors   []Pid
	monitorsMu sync.Mutex
	up         bool
	id         string
}

// Send sends data to an actor on the local system
func (p *LocalPid) Send(data interface{}) {
	select {
	case p.incoming <- data:
	default:
	}
}

// Receive receives data from the inbox of a local actor
func (p *LocalPid) Receive() interface{} {
	return <-p.incoming
}

// Monitors returns the monitors of a local actor
func (p *LocalPid) Monitors() []Pid {
	return p.monitors
}

// Up returns true if the actor pointed to by the PID is up
func (p *LocalPid) Up() bool {
	return p.up
}

// Down sets the up-state to false
func (p *LocalPid) Down() {
	p.up = false
}

// Monitor monitors an actor on the local system
func (p *LocalPid) Monitor(by Pid) {
	p.monitorsMu.Lock()
	defer p.monitorsMu.Unlock()

	p.monitors = append(p.monitors, by)
}
