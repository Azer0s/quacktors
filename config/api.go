package config

func SetLogger(l Logger) {
	logger = l
}

func GetLogger() Logger {
	return logger
}

func SetQpmdPort(port uint16) {
	qpmdPort = port
}

func GetQpmdPort() uint16 {
	return qpmdPort
}
