package config

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Init()
	Trace(string, ...interface{})
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Fatal(string, ...interface{})
}

type LogrusLogger struct {
	Log *logrus.Logger
}

func (l *LogrusLogger) Init() {
	l.Log = logrus.StandardLogger()
	l.Log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	l.Log.SetLevel(logrus.TraceLevel)
}

func (l *LogrusLogger) Trace(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.TraceLevel) {
		l.Log.WithFields(toMap(values...)).Trace(message)
	}
}

func (l *LogrusLogger) Debug(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.DebugLevel) {
		l.Log.WithFields(toMap(values...)).Debug(message)
	}
}

func (l *LogrusLogger) Info(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.InfoLevel) {
		l.Log.WithFields(toMap(values...)).Info(message)
	}
}

func (l *LogrusLogger) Warn(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.WarnLevel) {
		l.Log.WithFields(toMap(values...)).Warn(message)
	}
}
func (l *LogrusLogger) Error(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.ErrorLevel) {
		l.Log.WithFields(toMap(values...)).Error(message)
	}
}

func (l *LogrusLogger) Fatal(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.FatalLevel) {
		l.Log.WithFields(toMap(values...)).Fatal(message)
	}
}

func toMap(values ...interface{}) map[string]interface{} {
	if (len(values) % 2) != 0 {
		panic("invalid logging parameters")
	}

	vals := make(map[string]interface{})
	for i := 0; i < len(values); i += 2 {
		if k, ok := values[i].(string); ok {
			vals[k] = values[i+1]
			continue
		}
		panic("expected key to be a string")
	}

	return vals
}

type NoopLogger struct {
}

func (l *NoopLogger) Init() {
}

func (l *NoopLogger) Trace(message string, values ...interface{}) {
}

func (l *NoopLogger) Debug(message string, values ...interface{}) {
}

func (l *NoopLogger) Info(message string, values ...interface{}) {
}

func (l *NoopLogger) Warn(message string, values ...interface{}) {
}

func (l *NoopLogger) Error(message string, values ...interface{}) {
}

func (l *NoopLogger) Fatal(message string, values ...interface{}) {
}
