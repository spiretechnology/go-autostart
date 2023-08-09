package autostart

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/spiretechnology/go-autostart/internal/errutil"
)

//go:embed templates/linux_systemd.service
var systemdTemplate string

type systemdTemplateData struct {
	Options    Options
	BinaryPath string
	Arguments  string
	WantedBy   string
}

func (a *autostart) IsEnabled() (bool, error) {
	// Get the path to the systemd unit file
	systemdPath, err := a.getSystemdUnitFilePath()
	if err != nil {
		return false, errutil.Wrap(err)
	}

	// Check if the systemd unit file exists
	if _, err := os.Stat(systemdPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errutil.Wrapf(err, "checking systemd unit file")
	}

	// Check the status of the service
	statusCode, err := a.systemctlRun("status")
	if err != nil {
		return false, errutil.Wrapf(err, "checking systemd service status")
	}
	return statusCode == 0, nil
}

func (a *autostart) Enable() error {
	// Create the systemd unit file
	if err := a.createSystemdUnitFile(); err != nil {
		return errutil.Wrapf(err, "creating systemd unit file")
	}

	// Enable the service
	if _, err := a.systemctlRun("enable"); err != nil {
		return errutil.Wrapf(err, "enabling systemd service")
	}
	return nil
}

func (a *autostart) Disable() error {
	// Disable the service
	if _, err := a.systemctlRun("disable"); err != nil {
		return errutil.Wrapf(err, "disabling systemd service")
	}

	// Delete the systemd unit file if it exists
	systemdPath, err := a.getSystemdUnitFilePath()
	if err != nil {
		return errutil.Wrap(err)
	}
	if _, err := os.Stat(systemdPath); err == nil {
		if err := os.Remove(systemdPath); err != nil {
			return errutil.Wrapf(err, "deleting systemd unit file")
		}
	}
	return nil
}

func (a *autostart) getSystemdDir() (string, error) {
	switch a.options.Mode {
	case ModeUser:
		return filepath.Join(os.Getenv("HOME"), ".config", "systemd", "user"), nil
	case ModeSystem:
		return filepath.Join("/", "etc", "systemd", "system"), nil
	default:
		return "", ErrInvalidMode
	}
}

func (a *autostart) getServiceName() string {
	return fmt.Sprintf("%s.service", a.options.Label)
}

func (a *autostart) getSystemdUnitFilePath() (string, error) {
	systemdDir, err := a.getSystemdDir()
	if err != nil {
		return "", errutil.Wrap(err)
	}
	return filepath.Join(systemdDir, a.getServiceName()), nil
}

func (a *autostart) createSystemdUnitFile() error {
	// Get the path to the binary file
	binaryPath, err := os.Executable()
	if err != nil {
		return errutil.Wrapf(err, "getting executable path")
	}

	// Parse the systemd template
	tmpl, err := template.New("systemd").Parse(systemdTemplate)
	if err != nil {
		return errutil.Wrapf(err, "parsing systemd unit template")
	}

	// The path to the systemd unit file
	systemdPath, err := a.getSystemdUnitFilePath()
	if err != nil {
		return errutil.Wrap(err)
	}

	// Make the directory if it doesn't exist
	systemdDir := filepath.Dir(systemdPath)
	if _, err := os.Stat(systemdDir); os.IsNotExist(err) {
		if err := os.MkdirAll(systemdDir, 0777); err != nil {
			return errutil.Wrapf(err, "creating systemd directory")
		}
	}

	// Open the output file
	file, err := os.Create(systemdPath)
	if err != nil {
		return errutil.Wrapf(err, "creating systemd unit file")
	}
	defer file.Close()

	// Render into the file
	templateData := &systemdTemplateData{
		Options:    a.options,
		BinaryPath: binaryPath,
		Arguments:  joinArgs(a.options.Arguments),
		WantedBy:   "default.target",
	}
	if a.options.Mode == ModeSystem {
		templateData.WantedBy = "multi-user.target"
	}
	return errutil.Wrapf(tmpl.Execute(file, templateData), "rendering systemd template")
}

func (a *autostart) systemctlRun(action string) (int, error) {
	// Create the command, with either sudo or not depending on the startup mode
	var cmd *exec.Cmd
	switch a.options.Mode {
	case ModeUser:
		cmd = exec.Command("systemctl", action, a.getServiceName())
	case ModeSystem:
		cmd = exec.Command("sudo", "systemctl", action, a.getServiceName())
	default:
		return -1, ErrInvalidMode
	}

	// Set the standard output and error
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return -1, errutil.Wrapf(err, "running systemctl %s", action)
	}
	return cmd.ProcessState.ExitCode(), nil
}

func joinArgs(arguments []string) string {
	var escapedArgs []string
	for _, arg := range arguments {
		escapedArgs = append(escapedArgs, strconv.Quote(arg))
	}
	return strings.Join(escapedArgs, " ")
}
