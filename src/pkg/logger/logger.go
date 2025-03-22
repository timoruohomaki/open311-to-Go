package logger

import (
	"fmt"
	"io"
	"log/syslog"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger interface
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Close() error
}

// logrusLogger implements Logger interface using logrus
type logrusLogger struct {
	logger *logrus.Logger
	hooks  []io.Closer
}

// New creates a new logger
func New(cfg struct {
	Level          string `json:"level"`
	Format         string `json:"format"`
	ToSyslog       bool   `json:"toSyslog"`
	SyslogFacility string `json:"syslogFacility"`
	SyslogTag      string `json:"syslogTag"`
}) (Logger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %v", err)
	}
	logger.SetLevel(level)

	// Set log format
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
	}

	// By default, log to stdout
	logger.SetOutput(os.Stdout)

	// Create logger wrapper
	l := &logrusLogger{
		logger: logger,
		hooks:  make([]io.Closer, 0),
	}

	// Add syslog hook if configured
	if cfg.ToSyslog {
		// Parse facility
		var facility syslog.Priority
		switch cfg.SyslogFacility {
		case "local0":
			facility = syslog.LOG_LOCAL0
		case "local1":
			facility = syslog.LOG_LOCAL1
		case "local2":
			facility = syslog.LOG_LOCAL2
		case "local3":
			facility = syslog.LOG_LOCAL3
		case "local4":
			facility = syslog.LOG_LOCAL4
		case "local5":
			facility = syslog.LOG_LOCAL5
		case "local6":
			facility = syslog.LOG_LOCAL6
		case "local7":
			facility = syslog.LOG_LOCAL7
		default:
			facility = syslog.LOG_LOCAL0
		}

		// Set syslog tag
		tag := cfg.SyslogTag
		if tag == "" {
			tag = "open311api"
		}

		// Create syslog hook
		hook, err := logrus.NewSyslogHook("", "", facility, tag)
		if err != nil {
			return nil, fmt.Errorf("failed to create syslog hook: %v", err)
		}
		logger.AddHook(hook)

		// Add hook to the closer list if it implements io.Closer
		if closer, ok := hook.(io.Closer); ok {
			l.hooks = append(l.hooks, closer)
		}
	}

	return l, nil
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *logrusLogger) Close() error {
	for _, hook := range l.hooks {
		if err := hook.Close(); err != nil {
			return err
		}
	}
	return nil
}
