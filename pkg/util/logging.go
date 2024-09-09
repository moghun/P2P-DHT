package util

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type CustomFormatter struct {
	logrus.TextFormatter
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	message := entry.Message
	if strings.HasSuffix(message, "\n") {
		message = strings.TrimSuffix(message, "\n")
	}
	log := []byte(message + "\n")
	return log, nil
}

func SetupLogging(logFile string) {
	// Use the custom formatter instead of the default logrus formatter
	log.SetFormatter(&CustomFormatter{
		TextFormatter: logrus.TextFormatter{
			DisableTimestamp: true,
		},
	})

	// Open the log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// If there is an error opening the file, just log to stdout
		log.SetOutput(os.Stdout)
		return
	}

	// Create a multi-writer to write both to file and stdout
	multiWriter := io.MultiWriter(file, os.Stdout)
	log.SetOutput(multiWriter)
}

func Log() *logrus.Logger {
	return log
}
