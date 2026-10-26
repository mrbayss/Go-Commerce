package config

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// NewLogger creates a new logrus.Logger instance with JSON formatting.
func NewLogger(v *viper.Viper) *logrus.Logger {
	log := logrus.New()
	level, err := logrus.ParseLevel(v.GetString("log.level"))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	return log
}
