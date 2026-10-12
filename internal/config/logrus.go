package config

import (
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// NewLogger creates a new logrus.Logger instance with JSON formatting.
func NewLogger(viper *viper.Viper) *logrus.Logger {
	log := logrus.New()

	// Default log level
	level, err := logrus.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return f.Function, ""
		},
	})
	log.SetReportCaller(true)

	return log
}
