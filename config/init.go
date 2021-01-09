package config

import "github.com/Azer0s/quacktors/logging"

var logger logging.Logger
var qpmdPort uint16

func init() {
	logger = &logging.LogrusLogger{}
	logger.Init()
	qpmdPort = 7161
}
