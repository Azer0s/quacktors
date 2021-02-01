package quacktors

import (
	"github.com/Azer0s/quacktors/config"
	"github.com/Azer0s/quacktors/logging"
	"github.com/Azer0s/quacktors/metrics"
	"testing"
	"time"
)

func init() {
	config.SetLogger(&logging.NoopLogger{})
}

func TestSpawnMetric(t *testing.T) {
	metrics.RegisterRecorder(metrics.NewTimedRecorder(&metrics.TimedConsoleHook{}, 1*time.Second))

	for i := 0; i < 100_000; i++ {
		SpawnWithInit(func(ctx *Context) {
			ctx.SendAfter(ctx.Self(), PoisonPill{}, 500*time.Millisecond)
		}, func(ctx *Context, message Message) {

		})
	}

	<-time.After(1 * time.Second)

	Run()
}

func TestSendMetric(t *testing.T) {
	metrics.RegisterRecorder(metrics.NewTimedRecorder(&metrics.TimedConsoleHook{}, 1*time.Second))

	rootContext := RootContext()

	dummy := Spawn(func(ctx *Context, message Message) {
	})

	for i := 0; i < 100_000; i++ {
		rootContext.Send(dummy, EmptyMessage{})
	}

	<-time.After(1 * time.Second)

	rootContext.Send(dummy, PoisonPill{})

	Run()
}

func TestDropMetric(t *testing.T) {
	metrics.RegisterRecorder(metrics.NewTimedRecorder(&metrics.TimedConsoleHook{}, 1*time.Second))

	rootContext := RootContext()

	dummy := Spawn(func(ctx *Context, message Message) {
		<-time.After(10 * time.Millisecond)
	})

	for i := 0; i < 100_000; i++ {
		rootContext.Send(dummy, EmptyMessage{})
	}

	rootContext.Kill(dummy)

	<-time.After(1 * time.Second)

	Run()

	<-time.After(1 * time.Second)
}
