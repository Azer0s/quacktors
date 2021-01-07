package quacktors

type contextLogger struct {
	pid string
}

func (c *contextLogger) Trace(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Trace(message, context...)
		return
	}
	logger.Trace(message, append([]interface{}{"actor_pid", c.pid}, context...)...)
}

func (c *contextLogger) Debug(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Debug(message, context...)
		return
	}
	logger.Debug(message, append([]interface{}{"actor_pid", c.pid}, context...)...)
}

func (c *contextLogger) Info(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Info(message, context...)
		return
	}
	logger.Info(message, append([]interface{}{"actor_pid", c.pid}, context...)...)
}

func (c *contextLogger) Warn(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Warn(message, context...)
		return
	}
	logger.Warn(message, append([]interface{}{"actor_pid", c.pid}, context...)...)
}

func (c *contextLogger) Error(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Error(message, context...)
		return
	}

	logger.Error(message, append([]interface{}{"actor_pid", c.pid}, context...)...)
}

func (c *contextLogger) Fatal(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Fatal(message, context...)
		return
	}
	logger.Fatal(message, append([]interface{}{"actor_pid", c.pid}, context...)...)
}
