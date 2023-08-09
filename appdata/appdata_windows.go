package autostart

import (
	"os"
	"path/filepath"
)

const Logs = "Logs"

func SystemDataDir(vendor, product string) string {
	return filepath.Join(
		os.Getenv("ProgramData"),
		vendor,
		product,
	)
}

func UserDataDir(vendor, product string) string {
	return filepath.Join(
		os.Getenv("AppData"),
		vendor,
		product,
	)
}
