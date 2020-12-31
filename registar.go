package quacktors

import "sync"

var machineId = ""
var pidMap = make(map[string]*Pid)
var pidMapMu = &sync.RWMutex{}
var systemWg = &sync.WaitGroup{}

func init() {
	machineId = uuidString()
}

func registerPid(pid *Pid) {
	pidMapMu.Lock()
	defer pidMapMu.Unlock()

	pid.Id = uuidString()
	pidMap[pid.Id] = pid

	systemWg.Add(1)
}

func deletePid(pidId string) {
	pidMapMu.Lock()
	defer pidMapMu.Unlock()

	delete(pidMap, pidId)

	systemWg.Done()
}

func getByPidId(pidId string) (*Pid, bool) {
	pidMapMu.RLock()
	defer pidMapMu.RUnlock()

	v, ok := pidMap[pidId]

	return v, ok
}

func Wait() {
	systemWg.Wait()
}

func MachineId() string {
	return machineId
}
