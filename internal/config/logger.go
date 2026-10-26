package config

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger creates a new logrus.Logger instance with JSON formatting.
func NewLogger(cfg *Config) *logrus.Logger {
	log := logrus.New()
	level, err := logrus.ParseLevel(cfg.Log.Level)
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
