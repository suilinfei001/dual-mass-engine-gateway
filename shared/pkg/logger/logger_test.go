package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/quality-gateway/shared/pkg/logger"
)

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    logger.Level
		expected string
	}{
		{"DEBUG", logger.DebugLevel, "DEBUG"},
		{"INFO", logger.InfoLevel, "INFO"},
		{"WARN", logger.WarnLevel, "WARN"},
		{"ERROR", logger.ErrorLevel, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.level.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.level.String())
			}
		})
	}
}

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		Output:  &buf,
		Level:   logger.DebugLevel,
		Service: "test-service",
	})

	tests := []struct {
		name    string
		logFunc func(string, ...logger.Field)
		level   string
	}{
		{"Debug", log.Debug, "DEBUG"},
		{"Info", log.Info, "INFO"},
		{"Warn", log.Warn, "WARN"},
		{"Error", log.Error, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc("test message", logger.String("key", "value"))

			output := buf.String()
			if !strings.Contains(output, tt.level) {
				t.Errorf("expected output to contain %s, got %s", tt.level, output)
			}
			if !strings.Contains(output, "test-service") {
				t.Errorf("expected output to contain service name, got %s", output)
			}
			if !strings.Contains(output, "test message") {
				t.Errorf("expected output to contain message, got %s", output)
			}
			if !strings.Contains(output, "key=") {
				t.Errorf("expected output to contain field key, got %s", output)
			}
		})
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		Output: &buf,
		Level:  logger.WarnLevel,
	})

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("debug message should be filtered")
	}
	if strings.Contains(output, "info message") {
		t.Error("info message should be filtered")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("warn message should be present")
	}
	if !strings.Contains(output, "error message") {
		t.Error("error message should be present")
	}
}

func TestLoggerWithPrefix(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		Output: &buf,
		Level:  logger.InfoLevel,
	})

	prefixed := log.WithPrefix("[TEST] ")
	prefixed.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "[TEST]") {
		t.Errorf("expected output to contain prefix, got %s", output)
	}
}

func TestLoggerWithService(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		Output: &buf,
		Level:  logger.InfoLevel,
	})

	withService := log.WithService("new-service")
	withService.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "new-service") {
		t.Errorf("expected output to contain new service name, got %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("expected output to contain message, got %s", output)
	}
}

func TestFields(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		Output: &buf,
		Level:  logger.InfoLevel,
	})

	log.Info("test", logger.String("str", "value"), logger.Int("int", 42), logger.Err(nil))

	output := buf.String()
	if !strings.Contains(output, "str=") {
		t.Error("expected string field")
	}
	if !strings.Contains(output, "int=42") {
		t.Error("expected int field")
	}
	if !strings.Contains(output, "error=<nil>") {
		t.Error("expected nil error field")
	}
}

func TestErrField(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		Output: &buf,
		Level:  logger.InfoLevel,
	})

	log.Info("test", logger.Err(&testError{msg: "test error"}))

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Errorf("expected error message in output, got %s", output)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestDisableTime(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		Output:      &buf,
		Level:       logger.InfoLevel,
		DisableTime: true,
	})

	log.Info("test message")

	output := buf.String()
	// Check that it doesn't start with a timestamp
	if len(output) > 0 && output[0] >= '0' && output[0] <= '9' {
		t.Error("expected output without timestamp, but it starts with a digit")
	}
}
