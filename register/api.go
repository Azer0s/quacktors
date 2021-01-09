package register

import (
	"github.com/Azer0s/quacktors"
	"sync"
)

var register = make(map[string]*quacktors.Pid)
var registerMu = &sync.RWMutex{}

//ModifyUnsafe passes both the actual pid-register as well as
//the pid-register-mutex to a callback function so they can be
//modified directly. This should be used with caution!
func ModifyUnsafe(action func(register *map[string]*quacktors.Pid, mu *sync.RWMutex)) {
	action(&register, registerMu)
}

//SetPid associates a *quacktors.Pid with a name.
func SetPid(name string, pid *quacktors.Pid) {
	registerMu.Lock()
	defer registerMu.Unlock()

	register[name] = pid
}

//UsePid passes the *quacktors.Pid associated with a name
//to a callback function so it can be used safely.
func UsePid(name string, action func(pid *quacktors.Pid)) {
	registerMu.RLock()
	defer registerMu.RUnlock()

	action(register[name])
}

//ChangePid expects a *quacktors.Pid from a supplier function
//so it can safely change the pid-mapping from one *quacktors.Pid
//to another.
func ChangePid(name string, supplier func() *quacktors.Pid) {
	registerMu.Lock()
	defer registerMu.Unlock()

	register[name] = supplier()
}

//DeletePid deletes the association from a *quacktors.Pid to a name.
func DeletePid(name string) {
	registerMu.Lock()
	defer registerMu.Unlock()

	delete(register, name)
}
