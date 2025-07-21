package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func Init() {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)

	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	Logger.SetLevel(logrus.InfoLevel)
}

func LogInfo(message string, fields logrus.Fields) {
	Logger.WithFields(fields).Info(message)
}

func LogError(message string, err error, fields logrus.Fields) {
	Logger.WithFields(fields).WithError(err).Error(message)
}
