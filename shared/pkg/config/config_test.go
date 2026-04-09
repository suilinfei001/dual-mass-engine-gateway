package config_test

import (
	"os"
	"testing"

	"github.com/quality-gateway/shared/pkg/config"
)

func TestLoaderDefaults(t *testing.T) {
	loader := config.NewLoader().
		WithDefault("key1", "default1").
		WithDefault("key2", "default2")

	if v := loader.GetString("key1"); v != "default1" {
		t.Errorf("expected default1, got %s", v)
	}
	if v := loader.GetString("key2"); v != "default2" {
		t.Errorf("expected default2, got %s", v)
	}
}

func TestLoaderWithPrefix(t *testing.T) {
	loader := config.NewLoader().WithPrefix("TEST")

	// Set environment variable
	os.Setenv("TEST_HOST", "localhost")
	defer os.Unsetenv("TEST_HOST")

	if v := loader.GetString("host"); v != "localhost" {
		t.Errorf("expected localhost, got %s", v)
	}
}

func TestLoaderInt(t *testing.T) {
	loader := config.NewLoader().WithDefault("port", "8080")

	if v := loader.GetInt("port"); v != 8080 {
		t.Errorf("expected 8080, got %d", v)
	}

	loader2 := config.NewLoader().WithDefault("invalid", "not-a-number")
	if v := loader2.GetInt("invalid"); v != 0 {
		t.Errorf("expected 0 for invalid int, got %d", v)
	}
}

func TestLoaderBool(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"TRUE", "TRUE", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"false", "false", false},
		{"0", "0", false},
		{"no", "no", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := config.NewLoader().WithDefault("key", tt.value)
			if v := loader.GetBool("key"); v != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, v)
			}
		})
	}
}

func TestLoaderFloat64(t *testing.T) {
	loader := config.NewLoader().WithDefault("rate", "1.5")

	if v := loader.GetFloat64("rate"); v != 1.5 {
		t.Errorf("expected 1.5, got %f", v)
	}
}

func TestLoaderMustPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing required key")
		}
	}()

	loader := config.NewLoader()
	loader.MustGetString("missing")
}

func TestLoaderMustSuccess(t *testing.T) {
	loader := config.NewLoader().WithDefault("key", "value")
	if v := loader.MustGetString("key"); v != "value" {
		t.Errorf("expected value, got %s", v)
	}
}

func TestLoadFromEnvMap(t *testing.T) {
	loader := config.NewLoader()
	loader.LoadFromEnvMap(map[string]string{"key": "value"})

	if v := loader.GetString("key"); v != "value" {
		t.Errorf("expected value, got %s", v)
	}
}

func TestServerConfig(t *testing.T) {
	t.Run("Address with host", func(t *testing.T) {
		cfg := config.ServerConfig{Host: "localhost", Port: 8080}
		if addr := cfg.Address(); addr != "localhost:8080" {
			t.Errorf("expected localhost:8080, got %s", addr)
		}
	})

	t.Run("Address without host defaults to 0.0.0.0", func(t *testing.T) {
		cfg := config.ServerConfig{Port: 8080}
		if addr := cfg.Address(); addr != "0.0.0.0:8080" {
			t.Errorf("expected 0.0.0.0:8080, got %s", addr)
		}
	})

	t.Run("Validate invalid port", func(t *testing.T) {
		cfg := config.ServerConfig{Port: -1}
		if err := cfg.Validate(); err == nil {
			t.Error("expected error for invalid port")
		}
	})

	t.Run("Validate valid config", func(t *testing.T) {
		cfg := config.ServerConfig{Port: 8080}
		if err := cfg.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestDatabaseConfig(t *testing.T) {
	t.Run("DSN for mysql", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     3306,
			Database: "testdb",
			Username: "user",
			Password: "pass",
		}
		expected := "user:pass@tcp(localhost:3306)/testdb?parseTime=true&loc=Local"
		if dsn := cfg.DSN(); dsn != expected {
			t.Errorf("expected %s, got %s", expected, dsn)
		}
	})

	t.Run("Validate missing driver", func(t *testing.T) {
		cfg := config.DatabaseConfig{}
		if err := cfg.Validate(); err == nil {
			t.Error("expected error for missing driver")
		}
	})

	t.Run("Validate valid config", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     3306,
			Database: "testdb",
			Username: "user",
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestHTTPClientConfig(t *testing.T) {
	t.Run("Validate sets defaults", func(t *testing.T) {
		cfg := config.HTTPClientConfig{}
		if err := cfg.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg.Timeout != 30 {
			t.Errorf("expected default timeout 30, got %d", cfg.Timeout)
		}
		if cfg.MaxIdleConns != 100 {
			t.Errorf("expected default max idle conns 100, got %d", cfg.MaxIdleConns)
		}
	})
}
