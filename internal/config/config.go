package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	App           AppConfig      `envPrefix:"APP_"`
	Server        ServerConfig   `envPrefix:"SERVER_"`
	Database      DatabaseConfig `envPrefix:"DATABASE_"`
	Redis         RedisConfig    `envPrefix:"REDIS_"`
	JWT           JWTConfig      `envPrefix:"JWT_"`
	Runner        RunnerConfig   `envPrefix:"RUNNER_"`
	ServiceAPIKey string         `env:"SERVICE_API_KEY" envDefault:""`
}

type AppConfig struct {
	Env            string   `env:"ENV" envDefault:"development" validate:"required,oneof=development staging production test"`
	Name           string   `env:"NAME" envDefault:"wt-bot-ms-bots-v1" validate:"required"`
	Version        string   `env:"VERSION" envDefault:"0.1.0" validate:"required"`
	BaseURL        string   `env:"BASE_URL" envDefault:"http://localhost:8080" validate:"required,url"`
	PublicBaseURL  string   `env:"PUBLIC_BASE_URL" envDefault:"http://localhost:8080" validate:"required,url"`
	SwaggerURL     string   `env:"SWAGGER_URL" envDefault:"/swagger/doc.json" validate:"required"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000,http://localhost:5173"`
	EncryptionKey  string   `env:"ENCRYPTION_KEY" envDefault:""`
}

type ServerConfig struct {
	Addr            string        `env:"ADDR" envDefault:":8080" validate:"required"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"10s" validate:"required"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s" validate:"required"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s" validate:"required"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s" validate:"required"`
}

type DatabaseConfig struct {
	URL             string        `env:"URL" validate:"required,url"`
	MaxConns        int32         `env:"MAX_CONNS" envDefault:"10" validate:"gte=1"`
	MinConns        int32         `env:"MIN_CONNS" envDefault:"1" validate:"gte=0"`
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME" envDefault:"30m" validate:"required"`
	MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME" envDefault:"5m" validate:"required"`
}

type RedisConfig struct {
	Addr         string        `env:"ADDR" envDefault:"localhost:6379"`
	Password     string        `env:"PASSWORD" envDefault:""`
	DB           int           `env:"DB" envDefault:"0"`
	DialTimeout  time.Duration `env:"DIAL_TIMEOUT" envDefault:"5s" validate:"required"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"3s" validate:"required"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"3s" validate:"required"`
	PoolSize     int           `env:"POOL_SIZE" envDefault:"10" validate:"gte=1"`
}

type JWTConfig struct {
	Secret    string        `env:"SECRET" validate:"required,min=32"`
	Issuer    string        `env:"ISSUER" envDefault:"wt-bot-ms-users-v1" validate:"required"`
	AccessTTL time.Duration `env:"ACCESS_TTL" envDefault:"15m" validate:"required"`
}

type RunnerConfig struct {
	URL        string        `env:"URL" validate:"omitempty,url"`
	ServiceKey string        `env:"SERVICE_KEY" envDefault:""`
	Timeout    time.Duration `env:"TIMEOUT" envDefault:"10s" validate:"required"`
}

func Load() (Config, error) {
	return LoadFromPath(".env")
}

func LoadFromPath(envFile string) (Config, error) {
	_ = godotenv.Load(envFile)
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	if err := validate.Struct(cfg); err != nil {
		return Config{}, err
	}
	cfg.App.AllowedOrigins = normalizeCSV(cfg.App.AllowedOrigins)
	return cfg, nil
}

var validate = validator.New()

func normalizeCSV(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("load config: %v", err))
	}
	return cfg
}

func EnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
