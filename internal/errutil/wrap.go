package errutil

import "fmt"

func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("autostart: %w", err)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	if format == "" {
		return fmt.Errorf("autostart: %w", err)
	}
	return fmt.Errorf("autostart: %s: %w", fmt.Sprintf(format, args...), err)
}
