package config

import (
	"github.com/sirupsen/logrus"
	"os"
)

//The Logger interface is an abstraction for logging from quacktors.
//By default, a logrus implementation is included but another logger
//like zap or zerolog can easily be implemented.
type Logger interface {
	Init()
	Trace(string, ...interface{})
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Fatal(string, ...interface{})
}

//The LogrusLogger is the default logger for quacktors.
//As the name implies, it uses logrus under the hood.
type LogrusLogger struct {
	Log *logrus.Logger
}

//Init initializes the LogrusLogger with the default config (ForceColors=true, LogLevel=Trace)
func (l *LogrusLogger) Init() {
	l.Log = logrus.StandardLogger()
	l.Log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	l.Log.SetLevel(logrus.TraceLevel)
}

//Trace adds a logrus log entry on the corresponding log level
func (l *LogrusLogger) Trace(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.TraceLevel) {
		l.Log.WithFields(toMap(values...)).Trace(message)
	}
}

//Debug adds a logrus log entry on the corresponding log level
func (l *LogrusLogger) Debug(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.DebugLevel) {
		l.Log.WithFields(toMap(values...)).Debug(message)
	}
}

//Info adds a logrus log entry on the corresponding log level
func (l *LogrusLogger) Info(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.InfoLevel) {
		l.Log.WithFields(toMap(values...)).Info(message)
	}
}

//Warn adds a logrus log entry on the corresponding log level
func (l *LogrusLogger) Warn(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.WarnLevel) {
		l.Log.WithFields(toMap(values...)).Warn(message)
	}
}

//Error adds a logrus log entry on the corresponding log level
func (l *LogrusLogger) Error(message string, values ...interface{}) {
	if l.Log.IsLevelEnabled(logrus.ErrorLevel) {
		l.Log.WithFields(toMap(values...)).Error(message)
	}
}

//Fatal adds a logrus log entry on the corresponding log level
//and quits the application with exit-code 1
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

//The NoopLogger is an empty implementation of Logger.
//This can be used to disable the logging
//capabilities of quacktors.
type NoopLogger struct {
}

//Init noop implementation
func (l *NoopLogger) Init() {
}

//Trace noop implementation
func (l *NoopLogger) Trace(message string, values ...interface{}) {
}

//Debug noop implementation
func (l *NoopLogger) Debug(message string, values ...interface{}) {
}

//Info noop implementation
func (l *NoopLogger) Info(message string, values ...interface{}) {
}

//Warn noop implementation
func (l *NoopLogger) Warn(message string, values ...interface{}) {
}

//Error noop implementation
func (l *NoopLogger) Error(message string, values ...interface{}) {
}

//Fatal does not print (as this is a noop implementation), but does
//exit the application with exit code 1, as this is a "Fatal" method.
func (l *NoopLogger) Fatal(message string, values ...interface{}) {
	os.Exit(1)
}
