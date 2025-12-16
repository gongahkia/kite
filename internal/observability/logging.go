package observability

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog for structured logging
type Logger struct {
	logger zerolog.Logger
}

// NewLogger creates a new Logger
func NewLogger(level, format string) *Logger {
	var output io.Writer = os.Stdout

	// Set log level
	logLevel := parseLogLevel(level)
	zerolog.SetGlobalLevel(logLevel)

	// Set format
	if format == "text" || format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{
		logger: logger,
	}
}

// parseLogLevel parses log level string to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// ErrorWithErr logs an error with error object
func (l *Logger) ErrorWithErr(err error, msg string) {
	l.logger.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	logger := l.logger.With()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	return &Logger{
		logger: logger.Logger(),
	}
}

// WithContext adds request ID from context to logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return l.WithField("request_id", requestID)
	}
	return l
}

// GetZerologLogger returns the underlying zerolog.Logger
func (l *Logger) GetZerologLogger() zerolog.Logger {
	return l.logger
}

// SetGlobalLogger sets the global logger
func SetGlobalLogger(logger *Logger) {
	log.Logger = logger.logger
}

// RequestLogger creates a logger with request context
func RequestLogger(ctx context.Context, requestID, method, path string) *Logger {
	return &Logger{
		logger: log.With().
			Str("request_id", requestID).
			Str("method", method).
			Str("path", path).
			Logger(),
	}
}

// ScraperLogger creates a logger for scraping operations
func ScraperLogger(jurisdiction, source string) *Logger {
	return &Logger{
		logger: log.With().
			Str("jurisdiction", jurisdiction).
			Str("source", source).
			Str("component", "scraper").
			Logger(),
	}
}

// WorkerLogger creates a logger for worker operations
func WorkerLogger(workerID int, jobID string) *Logger {
	return &Logger{
		logger: log.With().
			Int("worker_id", workerID).
			Str("job_id", jobID).
			Str("component", "worker").
			Logger(),
	}
}
