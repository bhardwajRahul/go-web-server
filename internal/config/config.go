// Package config provides application configuration management using Viper.
package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
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
		MaxConnections  int           `mapstructure:"max_connections"`
		MinConnections  int           `mapstructure:"min_connections"`
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
}

// New creates and returns a new configuration instance with defaults, file, and environment overrides.
func New() *Config {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config file name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml") // Default, but will try others
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Try to read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Debug("no config file found, using defaults and environment variables")
		} else {
			slog.Warn("failed to read config file", "error", err)
		}
	} else {
		slog.Info("loaded configuration from file", "file", v.ConfigFileUsed())
	}

	// Environment variable handling
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		os.Exit(1)
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

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.host", "")
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.shutdown_timeout", 30*time.Second)

	// Database defaults
	v.SetDefault("database.url", "postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@localhost:5432/gowebserver?sslmode=disable")
	v.SetDefault("database.max_connections", 25)
	v.SetDefault("database.min_connections", 5)
	v.SetDefault("database.timeout", 30*time.Second)
	v.SetDefault("database.max_conn_lifetime", time.Hour)
	v.SetDefault("database.max_conn_idle_time", 30*time.Minute)
	v.SetDefault("database.run_migrations", true)
	v.SetDefault("database.ssl_mode", "disable")

	// Application defaults
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", false)
	v.SetDefault("app.log_level", "info")
	v.SetDefault("app.log_format", "text")

	// Security defaults
	v.SetDefault("security.trusted_proxies", []string{"127.0.0.1"})
	v.SetDefault("security.enable_cors", true)
	v.SetDefault("security.allowed_origins", []string{"*"})

	// Feature flags defaults
	v.SetDefault("features.enable_metrics", false)
	v.SetDefault("features.enable_pprof", false)
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
