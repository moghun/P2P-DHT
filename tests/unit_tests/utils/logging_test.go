package tests

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/DHT-14/pkg/util"
)

func TestSetupLoggingToFile(t *testing.T) {
	// Create a temporary log file
	tmpFile, err := os.CreateTemp("", "logfile.log")
	if err != nil {
		t.Fatalf("Failed to create temporary log file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up after the test

	// Set up logging
	util.SetupLogging(tmpFile.Name())

	// Log a test message
	testMessage := "This is a test log message"
	logrus.Info(testMessage)

	// Ensure log is flushed to the file
	logrus.StandardLogger().Out.(*os.File).Sync()

	// Read the log file
	logContent, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Check if the log file contains the test message
	if !bytes.Contains(logContent, []byte(testMessage)) {
		t.Errorf("Log file does not contain the expected message: %s", testMessage)
	}
}

func TestSetupLoggingToStdout(t *testing.T) {
	// Redirect stdout to a temporary file
	tmpFile, err := os.CreateTemp("", "stdout")
	if err != nil {
		t.Fatalf("Failed to create temporary file for stdout: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up after the test

	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }() // Restore original stdout
	os.Stdout = tmpFile

	// Set up logging with an invalid file path to fall back to stdout
	util.SetupLogging("/invalid/path/to/logfile.log")

	// Log a test message
	testMessage := "This is a test log message to stdout"
	logrus.Info(testMessage)

	// Ensure log is flushed to stdout
	logrus.StandardLogger().Out.(*os.File).Sync()

	// Read the stdout file
	tmpFile.Close() // Ensure the log is flushed to the file
	logContent, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read stdout file: %v", err)
	}

	// Check if the stdout file contains the test message
	if !bytes.Contains(logContent, []byte(testMessage)) {
		t.Errorf("Stdout does not contain the expected message: %s", testMessage)
	}
}
