package actors

import (
	"github.com/Azer0s/quacktors/pid"
	"github.com/Azer0s/quacktors/util"
	"sync"
)

var pidMap = make(map[uint64]pid.Pid)
var pidMapMu = &sync.RWMutex{}

// GetByGoid returns a PID by the goroutine ID
func GetByGoid(goid uint64) (pid.Pid, error) {
	pidMapMu.RLock()
	defer pidMapMu.RUnlock()

	if p, ok := pidMap[goid]; ok {
		return p, nil
	}

	return nil, util.PidDoesNotExistError()
}

// StoreByGoid stores a PID by its goroutine ID
func StoreByGoid(goid uint64, pid pid.Pid) {
	pidMapMu.Lock()
	defer pidMapMu.Unlock()

	pidMap[goid] = pid
}

// DeleteByGoid deletes a PID by its goroutine ID
func DeleteByGoid(goid uint64) {
	pidMapMu.Lock()
	defer pidMapMu.Unlock()

	delete(pidMap, goid)
}
