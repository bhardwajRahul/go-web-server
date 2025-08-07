// Package config provides application configuration management using Koanf.
package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

// Config holds all application configuration settings.
type Config struct {
	// Server configuration
	Server struct {
		Port            string        `koanf:"port"`
		Host            string        `koanf:"host"`
		ReadTimeout     time.Duration `koanf:"read_timeout"`
		WriteTimeout    time.Duration `koanf:"write_timeout"`
		ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`
	} `koanf:"server"`

	// Database configuration
	Database struct {
		URL             string        `koanf:"url"`
		MaxConnections  int           `koanf:"max_connections"`
		MinConnections  int           `koanf:"min_connections"`
		Timeout         time.Duration `koanf:"timeout"`
		MaxConnLifetime time.Duration `koanf:"max_conn_lifetime"`
		MaxConnIdleTime time.Duration `koanf:"max_conn_idle_time"`
		RunMigrations   bool          `koanf:"run_migrations"`
		SSLMode         string        `koanf:"ssl_mode"`
	} `koanf:"database"`

	// Application configuration
	App struct {
		Environment string `koanf:"environment"`
		Debug       bool   `koanf:"debug"`
		LogLevel    string `koanf:"log_level"`
		LogFormat   string `koanf:"log_format"`
	} `koanf:"app"`

	// Security configuration
	Security struct {
		TrustedProxies []string `koanf:"trusted_proxies"`
		EnableCORS     bool     `koanf:"enable_cors"`
		AllowedOrigins []string `koanf:"allowed_origins"`
	} `koanf:"security"`

	// Feature flags
	Features struct {
		EnableMetrics bool `koanf:"enable_metrics"`
		EnablePprof   bool `koanf:"enable_pprof"`
	} `koanf:"features"`
}

// New creates and returns a new configuration instance with defaults, file, and environment overrides.
func New() *Config {
	k := koanf.New(".")

	// Set defaults using structs provider
	cfg := getDefaults()
	if err := k.Load(structs.Provider(cfg, "koanf"), nil); err != nil {
		slog.Error("failed to load default config", "error", err)
		os.Exit(1)
	}

	// Load from config file if exists (supports JSON, YAML, TOML)
	configFiles := []struct {
		path   string
		parser koanf.Parser
	}{
		{"config.json", json.Parser()},
		{"config.yaml", yaml.Parser()},
		{"config.yml", yaml.Parser()},
		{"config.toml", toml.Parser()},
	}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile.path); err == nil {
			if err := k.Load(file.Provider(configFile.path), configFile.parser); err != nil {
				slog.Warn("failed to load config file", "file", configFile.path, "error", err)
			} else {
				slog.Info("loaded configuration from file", "file", configFile.path)

				break
			}
		}
	}

	// Load from environment variables (highest priority)
	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}), nil); err != nil {
		slog.Error("failed to load environment config", "error", err)
		os.Exit(1)
	}

	// Unmarshal into config struct
	var finalCfg Config
	if err := k.Unmarshal("", &finalCfg); err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		os.Exit(1)
	}

	// Production overrides
	if finalCfg.App.Environment == "production" {
		finalCfg.App.Debug = false
		finalCfg.App.LogFormat = "json"
		finalCfg.Security.AllowedOrigins = []string{}
		finalCfg.Database.RunMigrations = false
	}

	return &finalCfg
}

func getDefaults() Config {
	return Config{
		Server: struct {
			Port            string        `koanf:"port"`
			Host            string        `koanf:"host"`
			ReadTimeout     time.Duration `koanf:"read_timeout"`
			WriteTimeout    time.Duration `koanf:"write_timeout"`
			ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`
		}{
			Port:            "8080",
			Host:            "",
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: struct {
			URL             string        `koanf:"url"`
			MaxConnections  int           `koanf:"max_connections"`
			MinConnections  int           `koanf:"min_connections"`
			Timeout         time.Duration `koanf:"timeout"`
			MaxConnLifetime time.Duration `koanf:"max_conn_lifetime"`
			MaxConnIdleTime time.Duration `koanf:"max_conn_idle_time"`
			RunMigrations   bool          `koanf:"run_migrations"`
			SSLMode         string        `koanf:"ssl_mode"`
		}{
			URL:             "postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@localhost:5432/gowebserver?sslmode=disable",
			MaxConnections:  25,
			MinConnections:  5,
			Timeout:         30 * time.Second,
			MaxConnLifetime: time.Hour,
			MaxConnIdleTime: 30 * time.Minute,
			RunMigrations:   true,
			SSLMode:         "disable",
		},
		App: struct {
			Environment string `koanf:"environment"`
			Debug       bool   `koanf:"debug"`
			LogLevel    string `koanf:"log_level"`
			LogFormat   string `koanf:"log_format"`
		}{
			Environment: "development",
			Debug:       false,
			LogLevel:    "info",
			LogFormat:   "text",
		},
		Security: struct {
			TrustedProxies []string `koanf:"trusted_proxies"`
			EnableCORS     bool     `koanf:"enable_cors"`
			AllowedOrigins []string `koanf:"allowed_origins"`
		}{
			TrustedProxies: []string{"127.0.0.1"},
			EnableCORS:     true,
			AllowedOrigins: []string{"*"},
		},
		Features: struct {
			EnableMetrics bool `koanf:"enable_metrics"`
			EnablePprof   bool `koanf:"enable_pprof"`
		}{
			EnableMetrics: false,
			EnablePprof:   false,
		},
	}
}

// GetLogLevel converts the string log level to slog.Level.
func (c *Config) GetLogLevel() slog.Level {
	switch strings.ToLower(c.App.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
