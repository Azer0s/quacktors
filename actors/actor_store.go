package actors

import (
	"github.com/Azer0s/quacktors/pid"
	"github.com/Azer0s/quacktors/util"
	"sync"
)

var pidMap = make(map[uint64]pid.Pid)
var pidMapMu = &sync.RWMutex{}
var uniqueIdMap = make(map[string]uint64)
var uniqueIdMapMu = &sync.RWMutex{}

// GetByGoid returns a PID by the goroutine ID
func GetByGoid(goid uint64) (pid.Pid, error) {
	pidMapMu.RLock()
	defer pidMapMu.RUnlock()

	if p, ok := pidMap[goid]; ok {
		return p, nil
	}

	return nil, util.PidDoesNotExistError()
}

func GetByUniqueId(id string) (pid.Pid, error) {
	pidMapMu.RLock()
	defer pidMapMu.RUnlock()

	uniqueIdMapMu.RLock()
	defer uniqueIdMapMu.RUnlock()

	if goid, ok := uniqueIdMap[id]; ok {
		return pidMap[goid], nil
	}

	return nil, util.PidDoesNotExistError()
}

// StoreByGoid stores a PID by its goroutine ID
func StoreByGoid(goid uint64, pid pid.Pid, id string) {
	pidMapMu.Lock()
	defer pidMapMu.Unlock()

	pidMap[goid] = pid

	uniqueIdMapMu.Lock()
	defer uniqueIdMapMu.Unlock()
	uniqueIdMap[id] = goid
}

// DeleteByGoid deletes a PID by its goroutine ID
func DeleteByGoid(goid uint64, id string) {
	pidMapMu.Lock()
	defer pidMapMu.Unlock()
	delete(pidMap, goid)

	uniqueIdMapMu.Lock()
	defer uniqueIdMapMu.Unlock()
	delete(uniqueIdMap, id)
}
