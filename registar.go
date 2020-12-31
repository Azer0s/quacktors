package quacktors

import "sync"

var machineId = uuidString()
var pidMap = make(map[string]*Pid)
var pidMapMu = &sync.RWMutex{}
var systemWg = &sync.WaitGroup{}

var machines = map[string]*machine{}
var machinesMu = &sync.RWMutex{}

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

func registerMachine(machine *machine) {
	machinesMu.Lock()
	defer machinesMu.Unlock()

	machines[machine.machineId] = machine
}

func getMachine(machineId string) (*machine, bool) {
	machinesMu.RLock()
	defer machinesMu.RUnlock()

	v, ok := machines[machineId]

	return v, ok
}

func Wait() {
	systemWg.Wait()
}

func MachineId() string {
	return machineId
}
