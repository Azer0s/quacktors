package quacktors

import "sync"

var machineId = ""
var pidMap = make(map[string]*Pid)
var pidMapMu = &sync.RWMutex{}

func init() {
	machineId = uuidString()
}

func registerPid(pid *Pid) {
	pidMapMu.Lock()
	defer pidMapMu.Unlock()

	pid.Id = uuidString()
	pidMap[pid.Id] = pid
}

func deletePid(pidId string) {
	pidMapMu.Lock()
	defer pidMapMu.Unlock()

	delete(pidMap, pidId)
}

func getByPidId(pidId string) (*Pid, bool) {
	pidMapMu.RLock()
	defer pidMapMu.RUnlock()

	v, ok := pidMap[pidId]

	return v, ok
}

func MachineId() string {
	return machineId
}
