// Package config provides configuration management for all microservices.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config is the base configuration interface.
type Config interface {
	// Validate validates the configuration.
	Validate() error
	// ServiceName returns the service name.
	ServiceName() string
}

// Loader loads configuration from various sources.
type Loader struct {
	prefix      string
	defaults    map[string]string
	environment map[string]string
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	return &Loader{
		defaults:    make(map[string]string),
		environment: make(map[string]string),
	}
}

// WithPrefix sets a prefix for environment variable lookup.
func (l *Loader) WithPrefix(prefix string) *Loader {
	l.prefix = prefix
	return l
}

// WithDefault sets a default value for a key.
func (l *Loader) WithDefault(key, value string) *Loader {
	l.defaults[key] = value
	return l
}

// WithDefaults sets multiple default values.
func (l *Loader) WithDefaults(defaults map[string]string) *Loader {
	for k, v := range defaults {
		l.defaults[k] = v
	}
	return l
}

// GetString gets a string value.
func (l *Loader) GetString(key string) string {
	return l.get(key)
}

// GetInt gets an integer value.
func (l *Loader) GetInt(key string) int {
	v := l.get(key)
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return i
}

// GetInt64 gets an int64 value.
func (l *Loader) GetInt64(key string) int64 {
	v := l.get(key)
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

// GetBool gets a boolean value.
func (l *Loader) GetBool(key string) bool {
	v := strings.ToLower(l.get(key))
	return v == "true" || v == "1" || v == "yes"
}

// GetFloat64 gets a float64 value.
func (l *Loader) GetFloat64(key string) float64 {
	v := l.get(key)
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return f
}

// get is the internal method to get a value.
func (l *Loader) get(key string) string {
	// Check environment variable (with prefix if set)
	envKey := key
	if l.prefix != "" {
		envKey = l.prefix + "_" + strings.ToUpper(key)
	}
	if val := os.Getenv(envKey); val != "" {
		return val
	}

	// Check without prefix
	if l.prefix != "" {
		if val := os.Getenv(strings.ToUpper(key)); val != "" {
			return val
		}
	}

	// Check environment map
	if val, ok := l.environment[key]; ok {
		return val
	}

	// Check defaults
	if val, ok := l.defaults[key]; ok {
		return val
	}

	return ""
}

// MustGetString gets a string value or panics.
func (l *Loader) MustGetString(key string) string {
	v := l.GetString(key)
	if v == "" {
		panic(fmt.Sprintf("required config key %s is empty", key))
	}
	return v
}

// MustGetInt gets an integer value or panics.
func (l *Loader) MustGetInt(key string) int {
	v := l.GetInt(key)
	if v == 0 && l.get(key) == "" {
		panic(fmt.Sprintf("required config key %s is empty", key))
	}
	return v
}

// MustGetInt64 gets an int64 value or panics.
func (l *Loader) MustGetInt64(key string) int64 {
	v := l.GetInt64(key)
	if v == 0 && l.get(key) == "" {
		panic(fmt.Sprintf("required config key %s is empty", key))
	}
	return v
}

// MustGetBool gets a boolean value or panics.
func (l *Loader) MustGetBool(key string) bool {
	v := l.get(key)
	if v == "" {
		panic(fmt.Sprintf("required config key %s is empty", key))
	}
	return l.GetBool(key)
}

// LoadFromFile loads configuration from a JSON file.
func (l *Loader) LoadFromFile(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	return nil
}

// LoadFromEnvMap loads configuration from a map of environment variables.
func (l *Loader) LoadFromEnvMap(env map[string]string) {
	for k, v := range env {
		l.environment[k] = v
	}
}

// ServerConfig holds common server configuration.
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Address returns the server address.
func (c *ServerConfig) Address() string {
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Validate validates the server configuration.
func (c *ServerConfig) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	return nil
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Driver          string `json:"driver"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Database        string `json:"database"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
}

// DSN returns the data source name.
func (c *DatabaseConfig) DSN() string {
	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}

// Validate validates the database configuration.
func (c *DatabaseConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("database driver is required")
	}
	if c.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Port)
	}
	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Username == "" {
		return fmt.Errorf("database username is required")
	}
	return nil
}

// HTTPClientConfig holds HTTP client configuration.
type HTTPClientConfig struct {
	Timeout         int `json:"timeout"`
	MaxIdleConns    int `json:"max_idle_conns"`
	IdleConnTimeout int `json:"idle_conn_timeout"`
}

// Validate validates the HTTP client configuration.
func (c *HTTPClientConfig) Validate() error {
	if c.Timeout <= 0 {
		c.Timeout = 30 // default 30 seconds
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 100
	}
	if c.IdleConnTimeout <= 0 {
		c.IdleConnTimeout = 90
	}
	return nil
}

// LoadServerConfig loads server configuration from the loader.
func LoadServerConfig(loader *Loader, prefix string) ServerConfig {
	return ServerConfig{
		Host: loader.GetString(prefix + "_host"),
		Port: loader.GetInt(prefix + "_port"),
	}
}

// LoadDatabaseConfig loads database configuration from the loader.
func LoadDatabaseConfig(loader *Loader, prefix string) DatabaseConfig {
	return DatabaseConfig{
		Driver:          loader.GetString(prefix + "_driver"),
		Host:            loader.GetString(prefix + "_host"),
		Port:            loader.GetInt(prefix + "_port"),
		Database:        loader.GetString(prefix + "_database"),
		Username:        loader.GetString(prefix + "_username"),
		Password:        loader.GetString(prefix + "_password"),
		MaxOpenConns:    loader.GetInt(prefix + "_max_open_conns"),
		MaxIdleConns:    loader.GetInt(prefix + "_max_idle_conns"),
		ConnMaxLifetime: loader.GetInt(prefix + "_conn_max_lifetime"),
	}
}
