// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const iso8601 = "2006-01-02T15:04:05Z"

func newLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(utcFormatter{
		&logrus.TextFormatter{
			DisableColors:          true,
			DisableLevelTruncation: true,
			FullTimestamp:          true,
			TimestampFormat:        iso8601,
		},
	})
	return logger
}

type utcFormatter struct {
	logrus.Formatter
}

func (u utcFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	result, err := u.Formatter.Format(e)
	if err != nil {
		return nil, fmt.Errorf("formatting log message: %w", err)
	}
	return result, nil
}
