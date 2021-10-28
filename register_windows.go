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
)

func (app *App) getStartupDir() string {
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

func (app *App) getStartupLinkPath() string {
	return filepath.Join(
		app.getStartupDir(),
		fmt.Sprintf("%s.lnk", app.Name),
	)
}

func (app *App) IsRegistered() bool {
	if _, err := os.Stat(app.getStartupLinkPath()); err == nil {
		return true
	}
	return false
}

func (app *App) Register() error {

	// Get the path to the binary file. Also follow symbolic links to make sure we are always
	// using the absolute binary path and not the .lnk file
	binaryPath, err := os.Executable()
	if err != nil {
		return err
	}
	binaryPath, err = filepath.EvalSymlinks(binaryPath)
	if err != nil {
		return err
	}

	// Make the directory
	startupDir := app.getStartupDir()
	if err := os.MkdirAll(startupDir, 0777); err != nil {
		return err
	}

	// The path to the lnk file
	startupLnkPath := app.getStartupLinkPath()

	// If the binary path is ALREADY the symlink path, we don't need to do anything
	if binaryPath == startupLnkPath {
		return nil
	}

	// Create the shortcut using the Windows API
	args := ""
	res := C.CreateShortcut(C.CString(startupLnkPath), C.CString(binaryPath), C.CString(args))
	if res != 0 {
		return errors.New(fmt.Sprintf("autostart: cannot create shortcut '%s' error code: 0x%.8x", startupLnkPath, res))
	}
	return nil

}

func (app *App) Deregister() error {

	// Get the path to the lnk file
	startupLnkPath := app.getStartupLinkPath()

	// If the link file still exists
	if _, err := os.Stat(startupLnkPath); err == nil {
		return os.Remove(startupLnkPath)
	}

	// Otherwise just do nothing
	return nil

}
