package util

import (
	"github.com/Azer0s/quacktors/pid"
)

// PidToLocalPid converts a PID to a LocalPid type and returns it
func PidToLocalPid(p pid.Pid) *pid.LocalPid {
	var l interface{} = p
	return l.(*pid.LocalPid)
}
