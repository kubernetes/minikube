package kafka

// Logger is general logging interface that can be provided by popular logging
// frameworks.
//
// * https://github.com/go-kit/kit/tree/master/log
// * https://github.com/husio/log
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// nullLogger implements Logger interface, but discards all messages
type nullLogger struct {
}

func (nullLogger) Debug(msg string, args ...interface{}) {}
func (nullLogger) Info(msg string, args ...interface{})  {}
func (nullLogger) Warn(msg string, args ...interface{})  {}
func (nullLogger) Error(msg string, args ...interface{}) {}
