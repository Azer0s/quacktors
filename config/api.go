package config

import (
	"github.com/Azer0s/quacktors/encoding"
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

//SetEncoder sets the encoder implementation used by quacktors.
//(MsgpackEncoder by default)
func SetEncoder(e encoding.MessageEncoder) {
	encoder = e
}

//GetEncoder gets the configured encoder implementation.
func GetEncoder() encoding.MessageEncoder {
	return encoder
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
