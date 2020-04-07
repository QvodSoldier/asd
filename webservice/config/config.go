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

type asdConfig struct {
	websocket WSConfig
}

type WSConfig struct {
	ReadBufferSize   int //`envconfig:"default=4096"`
	WriteBufferSize  int //`envconfig:"default=4096"`
	HandShakeTimeOut int //`envconfig:"default=5"`
	PingInterval     int //`envconfig:"default=20"`
	IdleTimeout      int //`envconfig:"default=120"`
}

func LoadConfig() asdConfig {
	viper.SetDefault("LOG_LEVEL", "INFO")
	viper.AutomaticEnv()
	configLogger()
	// TODO: 加载配置文件的流程
	a := asdConfig{}
	return a
}
