package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

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
