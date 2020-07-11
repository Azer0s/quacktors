package pid

import "sync"

type LocalPid struct {
	incoming   chan interface{}
	monitors   []Pid
	monitorsMu sync.Mutex
	up         bool
}

func (p *LocalPid) Send(data interface{}) {
	select {
	case p.incoming <- data:
	default:
	}
}

func (p *LocalPid) Receive() interface{} {
	return <-p.incoming
}

func (p *LocalPid) Monitors() []Pid {
	return p.monitors
}

func (p *LocalPid) Up() bool {
	return p.up
}

func (p *LocalPid) Down() {
	p.up = false
}

func (p *LocalPid) Monitor(by Pid) {
	p.monitorsMu.Lock()
	defer p.monitorsMu.Unlock()

	p.monitors = append(p.monitors, by)
}
