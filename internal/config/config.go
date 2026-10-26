package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the typed application configuration (replaces raw viper access).
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Log      LogConfig      `mapstructure:"log"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Port        string `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Pass            string        `mapstructure:"pass"`
	DB              string        `mapstructure:"db"`
	SSLMode         string        `mapstructure:"sslmode"`
	AutoMigrate     bool          `mapstructure:"auto_migrate"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

type RedisConfig struct {
	Address     string        `mapstructure:"address"`
	Pass        string        `mapstructure:"pass"`
	DB          int           `mapstructure:"db"`
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
}

type AuthConfig struct {
	OTPTtl          time.Duration `mapstructure:"otp_ttl"`
	SignupRateLimit struct {
		Max    int           `mapstructure:"max"`
		Window time.Duration `mapstructure:"window"`
	} `mapstructure:"signup_rate_limit"`
	ResendCooldown time.Duration `mapstructure:"resend_cooldown"`
}

type SMTPConfig struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Username    string        `mapstructure:"username"`
	Password    string        `mapstructure:"password"`
	FromAddress string        `mapstructure:"from_address"`
	FromName    string        `mapstructure:"from_name"`
	TLSMode     string        `mapstructure:"tls_mode"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

// Load reads config.yaml (or path "."), applies env overrides, unmarshals into Config.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Defaults for zero/invalid values.
	if cfg.Postgres.MaxOpenConns <= 0 {
		cfg.Postgres.MaxOpenConns = 25
	}
	if cfg.Postgres.MaxIdleConns <= 0 {
		cfg.Postgres.MaxIdleConns = 10
	}
	if cfg.Postgres.ConnMaxLifetime <= 0 {
		cfg.Postgres.ConnMaxLifetime = 5 * time.Minute
	}
	if cfg.Postgres.ConnMaxIdleTime <= 0 {
		cfg.Postgres.ConnMaxIdleTime = 1 * time.Minute
	}
	if cfg.Auth.OTPTtl <= 0 {
		cfg.Auth.OTPTtl = 10 * time.Minute
	}
	if cfg.Auth.SignupRateLimit.Max <= 0 {
		cfg.Auth.SignupRateLimit.Max = 3
	}
	if cfg.Auth.SignupRateLimit.Window <= 0 {
		cfg.Auth.SignupRateLimit.Window = 10 * time.Minute
	}
	if cfg.Auth.ResendCooldown <= 0 {
		cfg.Auth.ResendCooldown = 60 * time.Second
	}
	if cfg.Redis.DialTimeout <= 0 {
		cfg.Redis.DialTimeout = 5 * time.Second
	}
	if cfg.SMTP.Timeout <= 0 {
		cfg.SMTP.Timeout = 10 * time.Second
	}
	return cfg, nil
}
