package actors

import (
	"github.com/Azer0s/Quacktors/quacktors/pid"
	"github.com/Azer0s/Quacktors/quacktors/util"
	"sync"
)

var pidMap = make(map[uint64]pid.Pid)
var pidMapMu = &sync.RWMutex{}

func GetByGoid(goid uint64) (pid.Pid, error) {
	pidMapMu.RLock()
	defer pidMapMu.RUnlock()

	if p, ok := pidMap[goid]; ok {
		return p, nil
	}

	return nil, util.PidDoesNotExistError()
}

func StoreByGoid(goid uint64, pid pid.Pid) {
	pidMapMu.Lock()
	defer pidMapMu.Unlock()

	pidMap[goid] = pid
}

func DeleteByGoid(goid uint64) {
	pidMapMu.Lock()
	defer pidMapMu.Unlock()

	delete(pidMap, goid)
}
