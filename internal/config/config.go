package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Queue         QueueConfig         `mapstructure:"queue"`
	Worker        WorkerConfig        `mapstructure:"worker"`
	Scraper       ScraperConfig       `mapstructure:"scraper"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Auth          AuthConfig          `mapstructure:"auth"`
	Security      SecurityConfig      `mapstructure:"security"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	EnableGRPC      bool          `mapstructure:"enable_grpc"`
	GRPCPort        int           `mapstructure:"grpc_port"`
	EnableGraphQL   bool          `mapstructure:"enable_graphql"`
	EnableWebSocket bool          `mapstructure:"enable_websocket"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"` // postgres, sqlite, mongodb
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Database        string        `mapstructure:"database"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// QueueConfig holds job queue configuration
type QueueConfig struct {
	Driver      string `mapstructure:"driver"` // nats, redis, memory
	URL         string `mapstructure:"url"`
	MaxRetries  int    `mapstructure:"max_retries"`
	RetryDelay  time.Duration `mapstructure:"retry_delay"`
}

// WorkerConfig holds worker pool configuration
type WorkerConfig struct {
	Count          int           `mapstructure:"count"`
	JobTimeout     time.Duration `mapstructure:"job_timeout"`
	ShutdownGrace  time.Duration `mapstructure:"shutdown_grace"`
}

// ScraperConfig holds scraping configuration
type ScraperConfig struct {
	UserAgent         string        `mapstructure:"user_agent"`
	RequestTimeout    time.Duration `mapstructure:"request_timeout"`
	MaxRetries        int           `mapstructure:"max_retries"`
	RateLimitPerMin   int           `mapstructure:"rate_limit_per_min"`
	RespectRobotsTxt  bool          `mapstructure:"respect_robots_txt"`
	EnableProxies     bool          `mapstructure:"enable_proxies"`
	ConcurrentLimit   int           `mapstructure:"concurrent_limit"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	LogLevel        string `mapstructure:"log_level"`
	LogFormat       string `mapstructure:"log_format"` // json, text
	MetricsEnabled  bool   `mapstructure:"metrics_enabled"`
	MetricsPort     int    `mapstructure:"metrics_port"`
	TracingEnabled  bool   `mapstructure:"tracing_enabled"`
	TracingEndpoint string `mapstructure:"tracing_endpoint"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret       string        `mapstructure:"jwt_secret"`
	JWTExpiration   time.Duration `mapstructure:"jwt_expiration"`
	APIKeyEnabled   bool          `mapstructure:"api_key_enabled"`
	RateLimitPerMin int           `mapstructure:"rate_limit_per_min"`
}

// SecurityConfig holds security configuration (for main.go)
type SecurityConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
	APIKeys       map[string]string
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// Read environment variables
	v.SetEnvPrefix("KITE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults and env vars
	}

	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.shutdown_timeout", "10s")
	v.SetDefault("server.enable_grpc", false)
	v.SetDefault("server.grpc_port", 9090)
	v.SetDefault("server.enable_graphql", false)
	v.SetDefault("server.enable_websocket", false)

	// Database defaults
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.database", "kite.db")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)

	// Queue defaults
	v.SetDefault("queue.driver", "memory")
	v.SetDefault("queue.max_retries", 3)
	v.SetDefault("queue.retry_delay", "5s")

	// Worker defaults
	v.SetDefault("worker.count", 4)
	v.SetDefault("worker.job_timeout", "5m")
	v.SetDefault("worker.shutdown_grace", "30s")

	// Scraper defaults
	v.SetDefault("scraper.user_agent", "Kite/4.0 (Legal Research Bot; +https://github.com/gongahkia/kite)")
	v.SetDefault("scraper.request_timeout", "30s")
	v.SetDefault("scraper.max_retries", 3)
	v.SetDefault("scraper.rate_limit_per_min", 20)
	v.SetDefault("scraper.respect_robots_txt", true)
	v.SetDefault("scraper.enable_proxies", false)
	v.SetDefault("scraper.concurrent_limit", 10)

	// Observability defaults
	v.SetDefault("observability.log_level", "info")
	v.SetDefault("observability.log_format", "json")
	v.SetDefault("observability.metrics_enabled", true)
	v.SetDefault("observability.metrics_port", 9091)
	v.SetDefault("observability.tracing_enabled", false)

	// Auth defaults
	v.SetDefault("auth.jwt_expiration", "24h")
	v.SetDefault("auth.api_key_enabled", true)
	v.SetDefault("auth.rate_limit_per_min", 100)

	// Security defaults
	v.SetDefault("security.jwt_secret", "change-this-secret-in-production")
	v.SetDefault("security.jwt_expiration", "24h")
}

// validate validates the configuration
func validate(cfg *Config) error {
	// Validate server config
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	// Validate worker config
	if cfg.Worker.Count < 1 {
		return fmt.Errorf("worker count must be at least 1")
	}

	// Validate scraper config
	if cfg.Scraper.RateLimitPerMin < 1 {
		return fmt.Errorf("scraper rate limit must be at least 1")
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
	}
	if !validLogLevels[cfg.Observability.LogLevel] {
		return fmt.Errorf("invalid log level: %s", cfg.Observability.LogLevel)
	}

	return nil
}
