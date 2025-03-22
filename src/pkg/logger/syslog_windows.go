// pkg/logger/syslog_windows.go
//go:build windows
// +build windows

package logger

import (
	"io"

	"github.com/sirupsen/logrus"
)

// addSyslogHook is a no-op on Windows
func addSyslogHook(logger *logrus.Logger, cfg Config, hooks *[]io.Closer) error {
	// Syslog is not supported on Windows
	// This is a no-op function
	return nil
}
