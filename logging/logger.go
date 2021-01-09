package logging

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
