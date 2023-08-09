package autostart

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spiretechnology/go-autostart/appdata"
	"github.com/spiretechnology/go-autostart/internal/errutil"
	"github.com/spiretechnology/go-autostart/internal/logutil"
)

// Autostart defines the interface for an autostart app.
type Autostart interface {
	// IsEnabled checks if the app is register to auto-start.
	IsEnabled() (bool, error)
	// Enable registers the app to auto-start.
	Enable() error
	// Disable unregisters the app from auto-start.
	Disable() error
	// Stdio overrides the global stdout and stderr with log files. It returns a Writer that
	// can be written to if you still want to log to the console as well as the log file.
	Stdio() (io.Writer, error)
	// DataDir gets the path to a sensible data directory where you might store app data.
	// If no log file paths were provided, logs will be stored in a subdirectory of this directory.
	DataDir() string
	// StdOutPath gets the path to the stdout log file.
	StdOutPath() string
	// StdErrPath gets the path to the stderr log file.
	StdErrPath() string
}

type Mode uint8

const (
	// ModeUser autostarts the application when the user logs in.
	ModeUser Mode = iota
	// ModeSystem autostarts the application when the system boots. This requires
	// administrator privileges.
	ModeSystem
)

type Options struct {
	// Label is the unique identifier for the app in reverse-DNS notation.
	Label string
	// Vendor is the name of the vendor of the app.
	Vendor string
	// Name is the human-readable name of the app.
	Name string
	// Description is the human-readable description of the app.
	Description string
	// Mode defines the startup mode of the app (system or user).
	Mode Mode
	// Arguments the arguments to pass to the app when it is started.
	Arguments []string
	// StdOutPath is the path to the log file for the process's stdout.
	StdOutPath string
	// StdErrPath is the path to the log file for the process's stderr.
	StdErrPath string
}

// New creates a new autostart instance.
// label is the unique identifier for the app, used by macOS only.
// name is the human-readable name of the app.
func New(options Options) Autostart {
	return &autostart{options}
}

type autostart struct {
	options Options
}

func (a *autostart) Stdio() (io.Writer, error) {
	// Create the stdout writer
	stdout, err := logutil.OverrideStdout(os.Stdout, a.StdOutPath())
	if err != nil {
		return nil, errutil.Wrapf(err, "wrapping stdout")
	}

	// Create the stderr writer
	if err := logutil.OverrideStderr(os.Stderr, a.StdErrPath()); err != nil {
		return nil, errutil.Wrapf(err, "wrapping stderr")
	}
	return stdout, nil
}

func (a *autostart) DataDir() string {
	if a.options.Mode == ModeSystem {
		return appdata.SystemDataDir(a.options.Vendor, a.options.Name)
	} else {
		return appdata.UserDataDir(a.options.Vendor, a.options.Name)
	}
}

func (a *autostart) StdOutPath() string {
	if a.options.StdOutPath != "" {
		return a.options.StdOutPath
	}
	return filepath.Join(a.DataDir(), appdata.Logs, "stdout.log")
}

func (a *autostart) StdErrPath() string {
	if a.options.StdErrPath != "" {
		return a.options.StdErrPath
	}
	return filepath.Join(a.DataDir(), appdata.Logs, "stderr.err")
}
