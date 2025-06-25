package logger

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerWithSyslogConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Syslog not supported on Windows")
	}

	cfg := Config{
		Level:          "info",
		Format:         "text",
		ToSyslog:       true,
		SyslogFacility: "local0",
		SyslogHost:     "localhost",
		SyslogPort:     "514",
		SyslogTag:      "test-logger",
	}

	log, err := New(cfg)
	assert.NoError(t, err, "Logger should be created without error when syslog is enabled")
	assert.NotNil(t, log)
}
