package circuitbreaker

// Logger required for logging in circuitbreaker
type Logger interface {
	Debug(args ...interface{}) //forgot what it returns :P
	Error(args ...interface{})
	Info(args ...interface{})
}

// NoopLogger is a default logger used in circuitbreaker it does not do anything
type NoopLogger struct{}

// Debug ...
func (NoopLogger) Debug(args ...interface{}) {}

// Error ...
func (NoopLogger) Error(args ...interface{}) {}

// Info ...
func (NoopLogger) Info(args ...interface{}) {}
