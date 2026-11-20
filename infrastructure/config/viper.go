package config

import (
	"errors"
	"strings"

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
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate fails fast on zero values that would otherwise silently break
// behavior (uploads rejected with 422, commission lost).
func (c *Config) validate() error {
	if c.Store.MaxLogoSize <= 0 {
		return errors.New("config: store.max_logo_size must be greater than 0")
	}
	if c.Store.MaxBannerSize <= 0 {
		return errors.New("config: store.max_banner_size must be greater than 0")
	}
	if c.Store.CommissionRate < 0 {
		return errors.New("config: store.commission_rate must not be negative")
	}
	if c.Product.MaxImageSize <= 0 {
		return errors.New("config: product.max_image_size must be greater than 0")
	}
	if c.Product.MaxImagesPerProduct <= 0 {
		return errors.New("config: product.max_images_per_product must be greater than 0")
	}
	return nil
}
