package quacktors

import "go.uber.org/zap"

var logger *zap.Logger

func init() {
	l, err := zap.NewProduction()

	if err != nil {
		panic(err)
	}

	logger = l
}
