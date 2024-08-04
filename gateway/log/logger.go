package log

import (
	logrus "github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

var DefaultLogger = Logger{logrus.StandardLogger()}
