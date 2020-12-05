package main

import (
	"github.com/sirupsen/logrus"
)

var logger = logrus.New() // type *logrus.Logger

func init() {
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat:        "2006-02-01 15:04:05",
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		// CallerPrettyfier: func(f *runtime.Frame) (string, string) {
		// 	return "", fmt.Sprintf("%s:%d", formatFilePath(f.File), f.Line)
		// },
	})
}
