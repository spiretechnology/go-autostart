package autostart

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed launchd_template.plist
var plistTemplate string

type plistTemplateData struct {
	Label      string
	BinaryPath string
	StdOutPath string
	StdErrPath string
}

func (app *App) getPlistDir() string {
	return filepath.Join(
		os.Getenv("HOME"),
		"Library",
		"LaunchAgents",
	)
}

func (app *App) getPlistFilePath() string {
	return filepath.Join(
		app.getPlistDir(),
		fmt.Sprintf("%s.plist", app.Label),
	)
}

func (app *App) IsRegistered() bool {
	if _, err := os.Stat(app.getPlistFilePath()); err == nil {
		return true
	}
	return false
}

func (app *App) Register() error {

	// Make the directory
	if err := os.MkdirAll(app.getPlistDir(), 0777); err != nil {
		return err
	}

	// The path to the plist file
	plistPath := app.getPlistFilePath()

	// Get the path to the binary file
	binaryPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Parse the plist template
	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return err
	}

	// Open the output file
	file, err := os.Create(plistPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Render into the file
	templateData := &plistTemplateData{
		Label:      app.Label,
		BinaryPath: binaryPath,
		StdOutPath: "/dev/null",
		StdErrPath: "/dev/null",
	}
	return tmpl.Execute(file, templateData)

}

func (app *App) Deregister() error {

	// Get the path to the plist file
	startupLnkPath := app.getPlistFilePath()

	// If the plist file exists, delete it
	if _, err := os.Stat(startupLnkPath); err == nil {
		return os.Remove(startupLnkPath)
	}

	// Otherwise just do nothing
	return nil

}
