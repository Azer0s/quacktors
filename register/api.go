package register

import (
	"github.com/Azer0s/quacktors"
	"sync"
)

var register = make(map[string]*quacktors.Pid)
var registerMu = &sync.RWMutex{}

func ModifyUnsafe(action func(register *map[string]*quacktors.Pid, mu *sync.RWMutex)) {
	action(&register, registerMu)
}

func SetPid(name string, pid *quacktors.Pid) {
	registerMu.Lock()
	defer registerMu.Unlock()

	register[name] = pid
}

func UsePid(name string, action func(pid *quacktors.Pid)) {
	registerMu.RLock()
	defer registerMu.RUnlock()

	action(register[name])
}

func ChangePid(name string, supplier func() *quacktors.Pid) {
	registerMu.Lock()
	defer registerMu.Unlock()

	register[name] = supplier()
}

func DeletePid(name string) {
	registerMu.Lock()
	defer registerMu.Unlock()

	delete(register, name)
}
