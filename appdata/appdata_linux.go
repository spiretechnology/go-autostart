package appdata

import (
	"os"
	"path/filepath"
)

const Logs = "logs"

func SystemDataDir(vendor, product string) string {
	return filepath.Join(
		"/etc",
		vendor,
		product,
	)
}

func UserDataDir(vendor, product string) string {
	return filepath.Join(
		os.Getenv("HOME"),
		".local",
		"share",
		vendor,
		product,
	)
}
