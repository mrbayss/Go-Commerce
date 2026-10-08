package config

import (
	"fmt"

	"github.com/spf13/viper"
	_ "gopkg.in/yaml.v3"
)

// NewViper creates a new Viper instance and reads the configuration file.
func NewViper() *viper.Viper {
	config := viper.New()
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.AddConfigPath("./")

	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error failed to get config file: %v\n", err))
	}

	return config
}
