package config

import (
	"time"
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
