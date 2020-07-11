package util

import (
	"github.com/Azer0s/quacktors/pid"
)

func PidToLocalPid(p pid.Pid) *pid.LocalPid {
	var l interface{} = p
	return l.(*pid.LocalPid)
}
