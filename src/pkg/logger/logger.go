package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"

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

// Config represents logger configuration
type Config struct {
	Level          string `json:"level"`
	Format         string `json:"format"`
	ApacheLogPath  string `json:"apacheLogPath"`
	ToSyslog       bool   `json:"toSyslog"`
	SyslogFacility string `json:"syslogFacility"`
	SyslogHost     string `json:"syslogHost"`
	SyslogPort     string `json:"syslogPort"`
	SyslogTag      string `json:"syslogTag"`
}

// New creates a new logger
func New(cfg Config) (Logger, error) {
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

	// Add syslog hook if configured and on a supported platform
	if cfg.ToSyslog && runtime.GOOS != "windows" {
		err := addSyslogHook(logger, cfg, &l.hooks)
		if err != nil {
			return nil, err
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
