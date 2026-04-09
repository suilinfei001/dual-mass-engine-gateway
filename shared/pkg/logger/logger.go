// Package logger provides structured logging for all microservices.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Level represents the log level.
type Level int

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in production.
	DebugLevel Level = iota
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly, it shouldn't generate any error-level logs.
	ErrorLevel
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return fmt.Sprintf("LEVEL(%d)", l)
	}
}

// Logger is a structured logger.
type Logger struct {
	mu          sync.Mutex
	out         io.Writer
	level       Level
	service     string
	prefix      string
	disableTime bool
}

// Config holds the logger configuration.
type Config struct {
	// Output is the writer where logs will be written. Defaults to os.Stdout.
	Output io.Writer
	// Level is the minimum log level to be written.
	Level Level
	// Service is the service name to include in logs.
	Service string
	// Prefix is a prefix to add to all log messages.
	Prefix string
	// DisableTime disables timestamp output.
	DisableTime bool
}

// New creates a new logger with the given configuration.
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	return &Logger{
		out:         cfg.Output,
		level:       cfg.Level,
		service:     cfg.Service,
		prefix:      cfg.Prefix,
		disableTime: cfg.DisableTime,
	}
}

// Default returns a logger with default settings.
func Default() *Logger {
	return New(Config{
		Output: os.Stdout,
		Level:  InfoLevel,
	})
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput sets the output writer.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// WithPrefix returns a new logger with the given prefix.
func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		out:         l.out,
		level:       l.level,
		service:     l.service,
		prefix:      l.prefix + prefix,
		disableTime: l.disableTime,
	}
}

// WithService returns a new logger with the given service name.
func (l *Logger) WithService(service string) *Logger {
	return &Logger{
		out:         l.out,
		level:       l.level,
		service:     service,
		prefix:      l.prefix,
		disableTime: l.disableTime,
	}
}

// Debug logs a message at DebugLevel.
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(DebugLevel, msg, fields...)
}

// Info logs a message at InfoLevel.
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(InfoLevel, msg, fields...)
}

// Warn logs a message at WarnLevel.
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(WarnLevel, msg, fields...)
}

// Error logs a message at ErrorLevel.
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(ErrorLevel, msg, fields...)
}

// Fatal logs a message at ErrorLevel and exits the program.
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.log(ErrorLevel, msg, fields...)
	os.Exit(1)
}

// log is the internal logging method.
func (l *Logger) log(level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	var ts string
	if !l.disableTime {
		ts = time.Now().Format("2006-01-02 15:04:05.000")
	}

	// Build log message
	buf := make([]byte, 0, 256)
	if ts != "" {
		buf = append(buf, ts...)
		buf = append(buf, ' ')
	}
	buf = append(buf, level.String()...)
	buf = append(buf, ' ')
	if l.service != "" {
		buf = append(buf, l.service...)
		buf = append(buf, ' ')
	}
	if l.prefix != "" {
		buf = append(buf, l.prefix...)
		buf = append(buf, ' ')
	}
	buf = append(buf, msg...)

	// Add fields
	for _, f := range fields {
		buf = append(buf, ' ')
		buf = append(buf, f.Key...)
		buf = append(buf, '=')
		buf = append(buf, f.Value...)
	}

	buf = append(buf, '\n')
	l.out.Write(buf)
}

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value string
}

// String creates a string field.
func String(key, value string) Field {
	return Field{Key: key, Value: fmt.Sprintf("%q", value)}
}

// Int creates an int field.
func Int(key string, value int) Field {
	return Field{Key: key, Value: fmt.Sprintf("%d", value)}
}

// Int64 creates an int64 field.
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: fmt.Sprintf("%d", value)}
}

// Err creates an error field.
func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: "<nil>"}
	}
	return Field{Key: "error", Value: fmt.Sprintf("%q", err.Error())}
}

// Any creates a field with any value formatted with %v.
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: fmt.Sprintf("%v", value)}
}

// StdLogger returns a standard library logger that writes to this logger at InfoLevel.
func (l *Logger) StdLogger() *log.Logger {
	return log.New(&logWriter{l}, "", 0)
}

type logWriter struct {
	*Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.Info(string(p))
	return len(p), nil
}

// Global default logger
var std = Default()

// SetDefault sets the default global logger.
func SetDefault(l *Logger) {
	std = l
}

// Debug logs a message at DebugLevel using the default logger.
func Debug(msg string, fields ...Field) {
	std.Debug(msg, fields...)
}

// Info logs a message at InfoLevel using the default logger.
func Info(msg string, fields ...Field) {
	std.Info(msg, fields...)
}

// Warn logs a message at WarnLevel using the default logger.
func Warn(msg string, fields ...Field) {
	std.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel using the default logger.
func Error(msg string, fields ...Field) {
	std.Error(msg, fields...)
}

// Fatal logs a message at ErrorLevel and exits using the default logger.
func Fatal(msg string, fields ...Field) {
	std.Fatal(msg, fields...)
}
