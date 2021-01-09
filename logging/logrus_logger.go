package logging

import "github.com/sirupsen/logrus"

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
