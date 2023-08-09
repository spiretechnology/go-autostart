package appdata

import (
	"os"
	"path/filepath"
)

const Logs = "Logs"

func SystemDataDir(vendor, product string) string {
	return filepath.Join(
		"/Library/Application Support",
		vendor,
		product,
	)
}

func UserDataDir(vendor, product string) string {
	return filepath.Join(
		os.Getenv("HOME"),
		"Library",
		"Application Support",
		vendor,
		product,
	)
}
