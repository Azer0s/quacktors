package metrics

import "fmt"

//TimedConsoleHook is a simple TimedRecorderHook that
//just prints to stdout.
type TimedConsoleHook struct {
}

//Record is the callback for the TimedRecorder.
func (t *TimedConsoleHook) Record(metrics TimedMetrics) {
	fmt.Println("Spawn:", metrics.SpawnCount)
	fmt.Println("Die:", metrics.DieCount)
	fmt.Println("Drop (local):", metrics.DropCount)
	fmt.Println("Drop (remote):", metrics.RemoteDropCount)
	fmt.Println("Unhandled:", metrics.UnhandledCount)
	fmt.Println("Receive (total):", metrics.ReceiveTotalCount)
	fmt.Println("Receive (remote):", metrics.ReceiveRemoteCount)
	fmt.Println("Send (local):", metrics.SendLocalCount)
	fmt.Println("Send (remote):", metrics.SendRemoteCount)
}
