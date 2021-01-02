package config

var logger Logger
var qpmdPort uint16

func init() {
	logger = &LogrusLogger{}
	logger.Init()
	qpmdPort = 7161
}
