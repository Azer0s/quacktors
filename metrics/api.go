package metrics

import "sync"

var recorders = make([]Recorder, 0)
var recordersMu = &sync.RWMutex{}

func RegisterRecorder(recorder Recorder) {
	recordersMu.Lock()
	defer recordersMu.Unlock()

	recorder.Init()

	recorders = append(recorders, recorder)
}

func RecordSpawn(pid string) {
	go func() {
		recordersMu.RLock()
		defer recordersMu.RUnlock()

		for _, r := range recorders {
			r.RecordSpawn(pid)
		}
	}()
}

func RecordDie(pid string) {
	go func() {
		recordersMu.RLock()
		defer recordersMu.RUnlock()

		for _, r := range recorders {
			r.RecordDie(pid)
		}
	}()
}

func RecordDrop(pid string, amount int) {
	go func() {
		recordersMu.RLock()
		defer recordersMu.RUnlock()

		for _, r := range recorders {
			r.RecordDrop(pid, amount)
		}
	}()
}

func RecordDropRemote(machine string, amount int) {
	go func() {
		recordersMu.RLock()
		defer recordersMu.RUnlock()

		for _, r := range recorders {
			r.RecordDropRemote(machine, amount)
		}
	}()
}

func RecordUnhandled(target string) {
	go func() {
		recordersMu.RLock()
		defer recordersMu.RUnlock()

		for _, r := range recorders {
			r.RecordUnhandled(target)
		}
	}()
}

func RecordReceive(pid string) {
	go func() {
		recordersMu.RLock()
		defer recordersMu.RUnlock()

		for _, r := range recorders {
			r.RecordReceive(pid)
		}
	}()
}

func RecordReceiveRemote(pid string) {
	go func() {
		recordersMu.RLock()
		defer recordersMu.RUnlock()

		for _, r := range recorders {
			r.RecordReceiveRemote(pid)
		}
	}()
}

func RecordSendLocal(target string) {
	go func() {
		recordersMu.RLock()
		defer recordersMu.RUnlock()

		for _, r := range recorders {
			r.RecordSendLocal(target)
		}
	}()
}

func RecordSendRemote(target string) {
	go func() {
		recordersMu.RLock()
		defer recordersMu.RUnlock()

		for _, r := range recorders {
			r.RecordSendRemote(target)
		}
	}()
}
