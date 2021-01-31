package config

import (
	"github.com/Azer0s/quacktors/encoding"
	"github.com/Azer0s/quacktors/logging"
)

var logger logging.Logger
var encoder encoding.MessageEncoder
var qpmdPort uint16

func init() {
	logger = &logging.LogrusLogger{}
	logger.Init()

	encoder = encoding.NewMsgpackEncoder()

	qpmdPort = 7161
}
