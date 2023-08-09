package autostart

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spiretechnology/go-autostart/v2/internal/errutil"
)

//go:embed templates/macos_launchd.plist
var plistTemplate string

type plistTemplateData struct {
	Options          Options
	BinaryPath       string
	EscapedArguments []string
}

func (a *autostart) IsEnabled() (bool, error) {
	plistPath, err := a.getPlistFilePath()
	if err != nil {
		return false, errutil.Wrapf(err, "getting plist file path")
	}
	if _, err = os.Stat(plistPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errutil.Wrapf(err, "checking if plist file exists")
	}
	return true, nil
}

func (a *autostart) Enable() error {
	// Get the path to the binary file
	binaryPath, err := os.Executable()
	if err != nil {
		return errutil.Wrapf(err, "getting executable path")
	}

	// Parse the plist template
	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return errutil.Wrapf(err, "parsing plist template")
	}

	// The path to the plist file
	plistPath, err := a.getPlistFilePath()
	if err != nil {
		return errutil.Wrap(err)
	}

	// Make the directory if it doesn't exist
	plistDir := filepath.Dir(plistPath)
	if _, err := os.Stat(plistDir); os.IsNotExist(err) {
		if err := os.MkdirAll(plistDir, 0777); err != nil {
			return errutil.Wrapf(err, "creating LaunchAgents directory")
		}
	}

	// Open the output file
	file, err := os.Create(plistPath)
	if err != nil {
		return errutil.Wrapf(err, "creating plist file")
	}
	defer file.Close()

	// Render into the file
	templateData := &plistTemplateData{
		Options:          a.options,
		BinaryPath:       binaryPath,
		EscapedArguments: escapeArgs(a.options.Arguments),
	}
	return errutil.Wrapf(tmpl.Execute(file, templateData), "rendering plist template")
}

func (a *autostart) Disable() error {
	// Get the path to the plist file
	plistPath, err := a.getPlistFilePath()
	if err != nil {
		return errutil.Wrap(err)
	}

	// If the plist file exists, delete it
	if _, err := os.Stat(plistPath); err == nil {
		return errutil.Wrapf(os.Remove(plistPath), "deleting plist file")
	}
	return nil
}

func (a *autostart) getPlistDir() (string, error) {
	switch a.options.Mode {
	case ModeUser:
		return filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents"), nil
	case ModeSystem:
		return filepath.Join("/", "Library", "LaunchAgents"), nil
	default:
		return "", ErrInvalidMode
	}
}

func (a *autostart) getPlistFilePath() (string, error) {
	launchAgentsDir, err := a.getPlistDir()
	if err != nil {
		return "", errutil.Wrap(err)
	}
	return filepath.Join(launchAgentsDir, fmt.Sprintf("%s.plist", a.options.Label)), nil
}

func escapeArgs(arguments []string) []string {
	output := make([]string, len(arguments))
	for i, arg := range arguments {
		output[i] = escapeXML(arg)
	}
	return output
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
