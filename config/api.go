package config

import (
	"github.com/Azer0s/quacktors/logging"
)

//SetLogger sets the Logger implementation used by quacktors.
//(LogrusLogger by default)
func SetLogger(l logging.Logger) {
	logger = l
}

//GetLogger gets the configured Logger implementation.
func GetLogger() logging.Logger {
	return logger
}

//SetQpmdPort sets the port quacktors uses to connect to
//local and remote qpmd instances. (7161 by default)
func SetQpmdPort(port uint16) {
	qpmdPort = port
}

//GetQpmdPort gets the configured qpmd port.
func GetQpmdPort() uint16 {
	return qpmdPort
}
