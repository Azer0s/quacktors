package metrics

import "fmt"

type TimedConsoleHook struct {
}

func (t *TimedConsoleHook) RecordSpawn(count int32) {
	fmt.Println("Spawn:", count)
}

func (t *TimedConsoleHook) RecordDie(count int32) {
	fmt.Println("Die:", count)
}

func (t *TimedConsoleHook) RecordDrop(count int32) {
	fmt.Println("Drop (local):", count)
}

func (t *TimedConsoleHook) RecordDropRemote(count int32) {
	fmt.Println("Drop (remote):", count)
}

func (t *TimedConsoleHook) RecordUnhandled(count int32) {
	fmt.Println("Unhandled:", count)
}

func (t *TimedConsoleHook) RecordReceiveTotal(count int32) {
	fmt.Println("Receive (total):", count)
}

func (t *TimedConsoleHook) RecordReceiveRemote(count int32) {
	fmt.Println("Receive (remote):", count)
}

func (t *TimedConsoleHook) RecordSendLocal(count int32) {
	fmt.Println("Send (local):", count)
}

func (t *TimedConsoleHook) RecordSendRemote(count int32) {
	fmt.Println("Send (remote):", count)
}
