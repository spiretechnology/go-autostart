package autostart

// #cgo LDFLAGS: -lole32 -luuid
/*
#define WIN32_LEAN_AND_MEAN
#include <stdint.h>
#include <windows.h>
uint64_t CreateShortcut(char *shortcutA, char *path, char *args);
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spiretechnology/go-autostart/internal/errutil"
	"golang.org/x/sys/windows/svc/mgr"
)

func (a *autostart) IsEnabled() (bool, error) {
	switch a.options.Mode {
	case ModeUser:
		return a.isEnabledUser()
	case ModeSystem:
		return a.isEnabledAdmin()
	default:
		return false, errutil.Wrap(ErrInvalidMode)
	}
}

func (a *autostart) Enable() error {
	switch a.options.Mode {
	case ModeUser:
		return a.enableUser()
	case ModeSystem:
		return a.enableAdmin()
	default:
		return ErrInvalidMode
	}
}

func (a *autostart) Disable() error {
	switch a.options.Mode {
	case ModeUser:
		return a.disableUser()
	case ModeSystem:
		return a.disableAdmin()
	default:
		return errutil.Wrap(ErrInvalidMode)
	}
}

func (a *autostart) getUserStartupDir() string {
	return filepath.Join(
		os.Getenv("USERPROFILE"),
		"AppData",
		"Roaming",
		"Microsoft",
		"Windows",
		"Start Menu",
		"Programs",
		"Startup",
	)
}

func (a *autostart) getUserStartupLinkPath() string {
	return filepath.Join(
		a.getUserStartupDir(),
		fmt.Sprintf("%s.lnk", a.options.Name),
	)
}

// isEnabledUser checks if the app is enabled for startup at the user level.
func (a *autostart) isEnabledUser() (bool, error) {
	_, err := os.Stat(a.getUserStartupLinkPath())
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errutil.Wrapf(err, "checking if plist file exists")
}

func (a *autostart) enableUser() error {
	// Get the path to the binary file. Also follow symbolic links to make sure we are always
	// using the absolute binary path and not the .lnk file
	binaryPath, err := os.Executable()
	if err != nil {
		return errutil.Wrapf(err, "getting executable path")
	}
	binaryPath, err = filepath.EvalSymlinks(binaryPath)
	if err != nil {
		return errutil.Wrapf(err, "evaluating symbolic links")
	}

	// Make the directory
	startupDir := a.getUserStartupDir()
	if err := os.MkdirAll(startupDir, 0777); err != nil {
		return errutil.Wrapf(err, "creating startup directory")
	}

	// If the binary path is ALREADY the symlink path, we don't need to do anything
	startupLnkPath := a.getUserStartupLinkPath()
	if binaryPath == startupLnkPath {
		return errutil.Wrapf(nil, "binary path is already the symlink path")
	}

	// Escape and join all the command line arguments
	args := joinArgs(a.options.Arguments)

	// Create the shortcut using the Windows API
	if res := C.CreateShortcut(C.CString(startupLnkPath), C.CString(binaryPath), C.CString(args)); res != 0 {
		return errors.New(fmt.Sprintf("autostart: cannot create shortcut '%s' error code: 0x%.8x", startupLnkPath, res))
	}
	return nil
}

func (a *autostart) disableUser() error {
	// Get the path to the lnk file
	startupLnkPath := a.getUserStartupLinkPath()

	// If the link file still exists
	if _, err := os.Stat(startupLnkPath); err == nil {
		return os.Remove(startupLnkPath)
	}

	// Otherwise just do nothing
	return nil
}

// isEnabledUser checks if the app is enabled for startup at the administrator level.
func (a *autostart) isEnabledAdmin() (bool, error) {
	// Connect to the Windows service manager
	m, err := mgr.Connect()
	if err != nil {
		return false, errutil.Wrapf(err, "connecting to service manager")
	}
	defer m.Disconnect()

	// Attempt to get the service
	s, err := m.OpenService(a.options.Label)
	if err != nil {
		// The service is not registered
		return false, nil
	}
	defer s.Close()

	// The service is registered
	return true, nil
}

func (a *autostart) enableAdmin() error {
	m, err := mgr.Connect()
	if err != nil {
		return errutil.Wrapf(err, "connecting to service manager")
	}
	defer m.Disconnect()

	// Configure the service
	config := mgr.Config{
		StartType:   mgr.StartAutomatic,
		DisplayName: a.options.Name,
		Description: a.options.Description,
	}

	// Get the path to the binary file
	binaryPath, err := os.Executable()
	if err != nil {
		return errutil.Wrapf(err, "getting executable path")
	}

	// Create the service
	s, err := m.CreateService(a.options.Label, binaryPath, config, a.options.Arguments...)
	if err != nil {
		return errutil.Wrapf(err, "creating service")
	}
	defer s.Close()

	// Start the service after system reboots
	err = s.SetRecoveryActions([]mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 60000},
	}, uint32(0))
	if err != nil {
		return errutil.Wrapf(err, "setting recovery actions")
	}
	return nil
}

func (a *autostart) disableAdmin() error {
	// Connect to the Windows service manager
	m, err := mgr.Connect()
	if err != nil {
		return errutil.Wrapf(err, "connecting to service manager")
	}
	defer m.Disconnect()

	// Attempt to get the service
	s, err := m.OpenService(a.options.Label)
	if err != nil {
		// The service is not registered
		return nil
	}
	defer s.Close()

	// Delete the service
	return errutil.Wrapf(s.Delete(), "deleting service")
}

func joinArgs(arguments []string) string {
	var escapedArgs []string
	for _, arg := range arguments {
		escapedArgs = append(escapedArgs, syscall.EscapeArg(arg))
	}
	return strings.Join(escapedArgs, " ")
}
