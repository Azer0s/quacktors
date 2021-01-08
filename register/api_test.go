package register

import (
	"github.com/Azer0s/quacktors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterPid(t *testing.T) {
	SetPid("foo", nil)

	assert.Nil(t, register["foo"])
}

func TestUsePid(t *testing.T) {
	SetPid("foo", nil)

	UsePid("foo", func(pid *quacktors.Pid) {
		assert.Nil(t, pid)
	})
}

func TestChangePid(t *testing.T) {
	SetPid("foo", nil)

	UsePid("foo", func(pid *quacktors.Pid) {
		assert.Nil(t, pid)
	})

	ChangePid("foo", func() *quacktors.Pid {
		return &quacktors.Pid{
			MachineId: "test",
			Id:        "test",
		}
	})

	UsePid("foo", func(pid *quacktors.Pid) {
		assert.Equal(t, "test", pid.Id)
		assert.Equal(t, "test", pid.MachineId)
	})
}
