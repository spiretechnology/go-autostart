package logutil

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spiretechnology/go-autostart/internal/errutil"
)

// OverrideStdout creates a wrapper around stdout that writes to both a log file and
// the original stdout. The Close() method on the returned io.WriteCloser will
// only close the log file.
func OverrideStdout(stdout *os.File, logFilePath string) (io.Writer, error) {
	// If the log file path is empty, just return stdout.
	// But we don't want to close the stdout file descriptor, so we wrap it in a NopCloser.
	if logFilePath == "" {
		return stdout, nil
	}

	// Make sure the log directory exists
	if err := os.MkdirAll(filepath.Dir(logFilePath), 0777); err != nil {
		return nil, errutil.Wrapf(err, "creating log directory")
	}

	// Create the log file
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errutil.Wrapf(err, "opening log file")
	}

	// Override the global stdout file descriptor with the log file.
	os.Stdout = logFile

	// Return a multi-write that allows us to write to both stdout and the log file.
	return io.MultiWriter(stdout, logFile), nil
}

// OverrideStderr overrides the global stderr file descriptor with a log file.
func OverrideStderr(stderr *os.File, logFilePath string) error {
	// If the log file path is empty, just return stdout.
	// But we don't want to close the stdout file descriptor, so we wrap it in a NopCloser.
	if logFilePath == "" {
		return nil
	}

	// Make sure the log directory exists
	if err := os.MkdirAll(filepath.Dir(logFilePath), 0777); err != nil {
		return errutil.Wrapf(err, "creating log directory")
	}

	// Create the log file
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errutil.Wrapf(err, "opening log file")
	}

	// Override the global stderr file descriptor with the log file
	// This ensures that any calls to log.Fatal() or panic() will write to the log file.
	os.Stderr = logFile
	return nil
}
