package quacktors

type contextLogger struct {
	pid string
}

func (c *contextLogger) Trace(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Trace(message, context...)
		return
	}
	logger.Trace(message, append(context, []interface{}{"actor_pid", c.pid})...)
}

func (c *contextLogger) Debug(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Debug(message, context...)
		return
	}
	logger.Debug(message, append(context, []interface{}{"actor_pid", c.pid})...)
}

func (c *contextLogger) Info(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Info(message, context...)
		return
	}
	logger.Info(message, append(context, []interface{}{"actor_pid", c.pid}...))
}

func (c *contextLogger) Warn(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Warn(message, context...)
		return
	}
	logger.Warn(message, append(context, []interface{}{"actor_pid", c.pid}...))
}

func (c *contextLogger) Error(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Error(message, context...)
		return
	}
	logger.Error(message, append(context, []interface{}{"actor_pid", c.pid})...)
}

func (c *contextLogger) Fatal(message string, context ...interface{}) {
	if c.pid == "root" {
		logger.Fatal(message, context...)
		return
	}
	logger.Fatal(message, append(context, []interface{}{"actor_pid", c.pid})...)
}
