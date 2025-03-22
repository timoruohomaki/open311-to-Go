// pkg/logger/syslog_unix.go
//go:build !windows
// +build !windows

package logger

import (
	"fmt"
	"io"
	"log/syslog"

	"github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
)

// addSyslogHook adds syslog hook to logger
func addSyslogHook(logger *logrus.Logger, cfg Config, hooks *[]io.Closer) error {
	// Parse facility
	var syslogFacility syslog.Priority
	switch cfg.SyslogFacility {
	case "local0":
		syslogFacility = syslog.LOG_LOCAL0
	case "local1":
		syslogFacility = syslog.LOG_LOCAL1
	case "local2":
		syslogFacility = syslog.LOG_LOCAL2
	case "local3":
		syslogFacility = syslog.LOG_LOCAL3
	case "local4":
		syslogFacility = syslog.LOG_LOCAL4
	case "local5":
		syslogFacility = syslog.LOG_LOCAL5
	case "local6":
		syslogFacility = syslog.LOG_LOCAL6
	case "local7":
		syslogFacility = syslog.LOG_LOCAL7
	default:
		syslogFacility = syslog.LOG_LOCAL0
	}

	// Set syslog tag
	tag := cfg.SyslogTag
	if tag == "" {
		tag = "go-api"
	}

	// Determine network and address
	network := ""
	address := ""

	if cfg.SyslogHost != "" {
		network = "udp"
		address = cfg.SyslogHost
		if cfg.SyslogPort != "" {
			address = fmt.Sprintf("%s:%s", cfg.SyslogHost, cfg.SyslogPort)
		}
	}

	// Create syslog hook
	hook, err := logrus_syslog.NewSyslogHook(network, address, syslogFacility, tag)
	if err != nil {
		return fmt.Errorf("failed to create syslog hook: %v", err)
	}
	logger.AddHook(hook)

	// Add hook to the closer list if it implements io.Closer
	if closer, ok := hook.(io.Closer); ok {
		*hooks = append(*hooks, closer)
	}

	return nil
}
