package config

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Logger = logrus.New()

func configLogger() {
	level := viper.GetString("LOG_LEVEL")
	switch level {
	case "DEBUG":
		Logger.SetLevel(logrus.DebugLevel)
	case "INFO":
		Logger.SetLevel(logrus.InfoLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}

	Logger.SetOutput(os.Stdout)
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	customFormatter.DisableColors = false
	Logger.SetFormatter(customFormatter)
}

func init() {
	viper.SetDefault("LOG_LEVEL", "INFO")
	viper.SetDefault("NAMESPACE", "default")
	viper.AutomaticEnv()
	configLogger()
}
