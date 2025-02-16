package logger

import (
	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	logger *logrus.Logger
}

func NewLogrusLogger() *LogrusLogger {
	log := logrus.New()
	log.SetFormatter(&ELKFormatter{})
	return &LogrusLogger{logger: log}
}

func (l *LogrusLogger) Info(msg string) {
	l.logger.Infoln(msg)
}

func (l *LogrusLogger) Error(msg string) {
	l.logger.Errorln(msg)
}
