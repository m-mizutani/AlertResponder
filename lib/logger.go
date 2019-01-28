package lib

import (
	"github.com/sirupsen/logrus"
)

// Logger is exported to allow replacement by external code.
var Logger = logrus.New()

func init() {
	Logger.SetLevel(logrus.DebugLevel)
	Logger.SetFormatter(&logrus.JSONFormatter{})
}
