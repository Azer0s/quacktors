package quacktors

type Abortable interface {
	Abort()
}

type MonitorAbortable struct {
	pid  *Pid
	self *Pid
}

func (ma *MonitorAbortable) Abort() {
	go func() {
		ma.pid.demonitorChanMu.RLock()
		defer ma.pid.demonitorChanMu.RUnlock()

		if ma.pid.demonitorChan == nil {
			return
		}

		ma.pid.demonitorChan <- ma.self
	}()
}

type SendAfterAbortable struct {
}

func (sa *SendAfterAbortable) Abort() {

}

type NoopAbortable struct {
}

func (na *NoopAbortable) Abort() {

}
