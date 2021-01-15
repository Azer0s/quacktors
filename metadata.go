package quacktors

import (
	"sync"
)

var remoteMonitorQuitAbortables = make(map[string]Abortable)
var remoteMonitorQuitAbortablesMu = &sync.RWMutex{}

var machineId = uuidString()
var pidMap = make(map[string]*Pid)
var pidMapMu = &sync.RWMutex{}
var systemWg = &sync.WaitGroup{}

var machines = map[string]*Machine{}
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

func registerMachine(machine *Machine) {
	machinesMu.Lock()
	defer machinesMu.Unlock()

	machines[machine.MachineId] = machine
}

func getMachine(machineId string) (*Machine, bool) {
	machinesMu.RLock()
	defer machinesMu.RUnlock()

	v, ok := machines[machineId]

	return v, ok
}

func deleteMachine(machineId string) {
	machinesMu.Lock()
	defer machinesMu.Unlock()

	delete(machines, machineId)
}

//Wait waits until all actors have quit.
func Run() {
	systemWg.Wait()
}

//MachineId returns the local machine id.
func MachineId() string {
	return machineId
}
