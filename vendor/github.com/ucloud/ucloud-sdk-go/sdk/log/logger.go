package log

import (
	"os"

	logrus "github.com/Sirupsen/logrus"
)

func Init(level logrus.Level) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(level)
	logrus.SetOutput(os.Stdout)
}
