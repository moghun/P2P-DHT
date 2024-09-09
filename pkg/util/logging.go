package util

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type CustomFormatter struct {
	logrus.TextFormatter
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Create a custom format, stripping "time", "level", and "msg" labels
	// and removing any trailing newlines from the message
	message := entry.Message
	if strings.HasSuffix(message, "\n") {
		message = strings.TrimSuffix(message, "\n")
	}

	// Format the message without additional labels
	log := []byte(message + "\n")
	return log, nil
}

func SetupLogging(logFile string) {
	// Use the custom formatter instead of the default logrus formatter
	logrus.SetFormatter(&CustomFormatter{
		TextFormatter: logrus.TextFormatter{
			DisableTimestamp: true, // Remove the timestamp from log output
		},
	})

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.SetOutput(os.Stdout)
	}
}

func Log() *logrus.Logger {
	return log
}