package node

import (
	"github.com/Azer0s/quacktors/util"
	"sync"
)

var systems = make(map[string]System)
var systemsMu = &sync.RWMutex{}

var systemPorts = make(map[string]int)
var systemPortsMu = &sync.RWMutex{}

func StoreSystem(system System) {
	systemsMu.Lock()
	defer systemsMu.Unlock()

	systems[system.name] = system
}

func StorePortBinding(port int, system string) {
	systemPortsMu.Lock()
	defer systemPortsMu.Unlock()

	systemPorts[system] = port
}

func GetSystem(system string) (System, error) {
	systemsMu.RLock()
	defer systemsMu.RUnlock()

	if s, ok := systems[system]; ok {
		return s, nil
	}

	return System{}, util.SystemDoesNotExistError()
}

func GetSystemPort(system string) (int, error) {
	systemPortsMu.RLock()
	defer systemPortsMu.RUnlock()

	if p, ok := systemPorts[system]; ok {
		return p, nil
	}

	return 0, util.SystemDoesNotExistError()
}
