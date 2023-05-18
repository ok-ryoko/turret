package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func newLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(UTCFormatter{
		&logrus.TextFormatter{
			DisableColors:          true,
			DisableLevelTruncation: true,
			FullTimestamp:          true,
			TimestampFormat:        ISO8601,
		},
	})
	return logger
}

const ISO8601 = "2006-01-02T15:04:05Z"

type UTCFormatter struct {
	logrus.Formatter
}

func (u UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	result, err := u.Formatter.Format(e)
	if err != nil {
		return []byte{}, fmt.Errorf("formatting log message: %w", err)
	}
	return result, nil
}
