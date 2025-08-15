// Package config provides application configuration management using koanf.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config holds all application configuration settings.
type Config struct {
	// Server configuration
	Server struct {
		Port            string        `mapstructure:"port"`
		Host            string        `mapstructure:"host"`
		ReadTimeout     time.Duration `mapstructure:"read_timeout"`
		WriteTimeout    time.Duration `mapstructure:"write_timeout"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	} `mapstructure:"server"`

	// Database configuration
	Database struct {
		URL             string        `mapstructure:"url"`
		MaxConnections  int32         `mapstructure:"max_connections"`
		MinConnections  int32         `mapstructure:"min_connections"`
		Timeout         time.Duration `mapstructure:"timeout"`
		MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
		MaxConnIdleTime time.Duration `mapstructure:"max_conn_idle_time"`
		RunMigrations   bool          `mapstructure:"run_migrations"`
		SSLMode         string        `mapstructure:"ssl_mode"`
	} `mapstructure:"database"`

	// Application configuration
	App struct {
		Environment string `mapstructure:"environment"`
		Debug       bool   `mapstructure:"debug"`
		LogLevel    string `mapstructure:"log_level"`
		LogFormat   string `mapstructure:"log_format"`
	} `mapstructure:"app"`

	// Security configuration
	Security struct {
		TrustedProxies []string `mapstructure:"trusted_proxies"`
		EnableCORS     bool     `mapstructure:"enable_cors"`
		AllowedOrigins []string `mapstructure:"allowed_origins"`
	} `mapstructure:"security"`

	// Feature flags
	Features struct {
		EnableMetrics bool `mapstructure:"enable_metrics"`
		EnablePprof   bool `mapstructure:"enable_pprof"`
	} `mapstructure:"features"`

	// JWT/Authentication configuration
	Auth struct {
		JWTSecret       string        `mapstructure:"jwt_secret"`
		TokenDuration   time.Duration `mapstructure:"token_duration"`
		RefreshDuration time.Duration `mapstructure:"refresh_duration"`
		CookieName      string        `mapstructure:"cookie_name"`
		CookieSecure    bool          `mapstructure:"cookie_secure"`
	} `mapstructure:"auth"`
}

// New creates and returns a new configuration instance with defaults, file, and environment overrides.
func New() *Config {
	k := koanf.New(".")

	// Set defaults
	setDefaults(k)

	// Try to read .env file first
	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		slog.Debug("no .env file found")
	} else {
		slog.Debug("loaded .env file")
	}

	// Try to read config file (optional)
	configFiles := []string{"config.yaml", "config.yml", "./config/config.yaml", "./config/config.yml"}
	configLoaded := false
	for _, configFile := range configFiles {
		if err := k.Load(file.Provider(configFile), yaml.Parser()); err == nil {
			slog.Info("loaded configuration from file", "file", configFile)
			configLoaded = true
			break
		}
	}
	if !configLoaded {
		slog.Debug("no config file found, using defaults and environment variables")
	}

	// Environment variable handling
	k.Load(env.Provider("", ".", func(s string) string {
		// Convert environment variables to lowercase and replace _ with .
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}), nil)

	// Unmarshal into config struct
	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "mapstructure"}); err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		os.Exit(1)
	}

	// Construct database URL if not provided directly
	if cfg.Database.URL == "" {
		user := k.String("database.user")
		password := k.String("database.password")
		host := k.String("database.host")
		port := k.String("database.port")
		name := k.String("database.name")
		sslmode := k.String("database.sslmode")

		// Set defaults for missing values
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}
		if name == "" {
			name = "gowebserver"
		}
		if sslmode == "" {
			sslmode = "disable"
		}

		if user != "" && password != "" {
			cfg.Database.URL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, name, sslmode)
		} else {
			slog.Error("DATABASE_URL not provided and DATABASE_USER/DATABASE_PASSWORD not found in environment")
			os.Exit(1)
		}
	}

	// Production overrides
	if cfg.App.Environment == "production" {
		cfg.App.Debug = false
		cfg.App.LogFormat = "json"
		cfg.Security.AllowedOrigins = []string{}
		cfg.Database.RunMigrations = false
	}

	return &cfg
}

func setDefaults(k *koanf.Koanf) {
	// Create a defaults map
	defaults := map[string]interface{}{
		// Server defaults
		"server.port":             "8080",
		"server.host":             "",
		"server.read_timeout":     10 * time.Second,
		"server.write_timeout":    10 * time.Second,
		"server.shutdown_timeout": 30 * time.Second,

		// Database defaults - will be overridden by environment variables
		"database.url":                "", // Will be constructed from individual vars if not set
		"database.max_connections":    25,
		"database.min_connections":    5,
		"database.timeout":            30 * time.Second,
		"database.max_conn_lifetime":  time.Hour,
		"database.max_conn_idle_time": 30 * time.Minute,
		"database.run_migrations":     true,
		"database.ssl_mode":           "disable",

		// Application defaults
		"app.environment": "development",
		"app.debug":       false,
		"app.log_level":   "info",
		"app.log_format":  "text",

		// Security defaults
		"security.trusted_proxies": []string{"127.0.0.1"},
		"security.enable_cors":     true,
		"security.allowed_origins": []string{"*"},

		// Feature flags defaults
		"features.enable_metrics": false,
		"features.enable_pprof":   false,

		// Authentication defaults
		"auth.jwt_secret":       "change-this-in-production",
		"auth.token_duration":   24 * time.Hour,
		"auth.refresh_duration": 7 * 24 * time.Hour,
		"auth.cookie_name":      "auth_token",
		"auth.cookie_secure":    true,
	}

	// Load defaults using the confmap provider
	k.Load(confmap.Provider(defaults, "."), nil)
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
