package logging

import (
	"log/slog"
	"os"
)

// InitLogger initializes structured logging with slog.
// Returns a logger configured for production or development based on env.
func InitLogger(dev bool) *slog.Logger {
	var opts *slog.HandlerOptions
	if dev {
		opts = &slog.HandlerOptions{Level: slog.LevelDebug}
	} else {
		opts = &slog.HandlerOptions{Level: slog.LevelInfo}
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

// ContextLogger wraps slog.Logger for convenient context logging.
type ContextLogger struct {
	logger *slog.Logger
}

// NewContextLogger creates a new ContextLogger.
func NewContextLogger(logger *slog.Logger) *ContextLogger {
	return &ContextLogger{logger: logger}
}

// Debug logs a debug message with key-value pairs.
func (cl *ContextLogger) Debug(msg string, args ...any) {
	cl.logger.Debug(msg, args...)
}

// Info logs an info message with key-value pairs.
func (cl *ContextLogger) Info(msg string, args ...any) {
	cl.logger.Info(msg, args...)
}

// Warn logs a warning message with key-value pairs.
func (cl *ContextLogger) Warn(msg string, args ...any) {
	cl.logger.Warn(msg, args...)
}

// Error logs an error message with key-value pairs.
func (cl *ContextLogger) Error(msg string, args ...any) {
	cl.logger.Error(msg, args...)
}

// WithValues returns a new logger with additional context values.
func (cl *ContextLogger) WithValues(args ...any) *ContextLogger {
	return &ContextLogger{logger: cl.logger.With(args...)}
}
