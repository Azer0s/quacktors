package logging

import "os"

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
