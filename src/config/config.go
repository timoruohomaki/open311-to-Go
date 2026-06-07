package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MongoDBConfig holds the MongoDB connection settings.
type MongoDBConfig struct {
	URI              string
	Database         string
	Collection       string
	ConnectTimeout   int
	OperationTimeout int
	// TLSCertificateKeyFile is the path to the PEM file containing the client
	// certificate and private key, used for MONGODB-X509 authentication.
	TLSCertificateKeyFile string
	// TLSCAFile is an optional path to a CA bundle. Leave empty to use the
	// system root CAs (sufficient for MongoDB Atlas).
	TLSCAFile string
}

// Config represents the application configuration. It is populated entirely from
// environment variables (see Load); there is no config file. The Logger struct
// must stay structurally identical to logger.Config so cfg.Logger can be passed
// to logger.New.
type Config struct {
	Server struct {
		Port                   int `json:"port"`
		ReadTimeoutSeconds     int `json:"readTimeoutSeconds"`
		WriteTimeoutSeconds    int `json:"writeTimeoutSeconds"`
		IdleTimeoutSeconds     int `json:"idleTimeoutSeconds"`
		ShutdownTimeoutSeconds int `json:"shutdownTimeoutSeconds"`
	}
	Logger struct {
		Level          string `json:"level"`
		Format         string `json:"format"`
		ApacheLogPath  string `json:"apacheLogPath"`
		ToSyslog       bool   `json:"toSyslog"`
		SyslogFacility string `json:"syslogFacility"`
		SyslogHost     string `json:"syslogHost"`
		SyslogPort     string `json:"syslogPort"`
		SyslogTag      string `json:"syslogTag"`
	}
	MongoDB MongoDBConfig
	Sentry  struct {
		DSN              string
		Environment      string
		EnableTracing    bool
		TracesSampleRate float64
		SendDefaultPII   bool
	}
}

// Load builds the configuration from environment variables, applying sensible
// defaults. Only MONGODB_URI is required. Call LoadDotEnv first if you want to
// source a local .env file.
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.Server.Port = getEnvInt("PORT", 8080)
	cfg.Server.ReadTimeoutSeconds = getEnvInt("READ_TIMEOUT_SECONDS", 30)
	cfg.Server.WriteTimeoutSeconds = getEnvInt("WRITE_TIMEOUT_SECONDS", 30)
	cfg.Server.IdleTimeoutSeconds = getEnvInt("IDLE_TIMEOUT_SECONDS", 120)
	cfg.Server.ShutdownTimeoutSeconds = getEnvInt("SHUTDOWN_TIMEOUT_SECONDS", 30)

	cfg.MongoDB.URI = getEnv("MONGODB_URI", "")
	cfg.MongoDB.Database = getEnv("MONGODB_DATABASE", "open311")
	cfg.MongoDB.Collection = getEnv("MONGODB_COLLECTION", "service_requests")
	cfg.MongoDB.ConnectTimeout = getEnvInt("MONGODB_CONNECT_TIMEOUT_SECONDS", 10)
	cfg.MongoDB.OperationTimeout = getEnvInt("MONGODB_OPERATION_TIMEOUT_SECONDS", 5)
	cfg.MongoDB.TLSCertificateKeyFile = getEnv("MONGODB_TLS_CERT_KEY_FILE", "")
	cfg.MongoDB.TLSCAFile = getEnv("MONGODB_TLS_CA_FILE", "")

	cfg.Logger.Level = getEnv("LOG_LEVEL", "info")
	cfg.Logger.Format = getEnv("LOG_FORMAT", "text")
	cfg.Logger.ApacheLogPath = getEnv("LOG_APACHE_PATH", "")
	cfg.Logger.ToSyslog = getEnvBool("LOG_TO_SYSLOG", false)
	cfg.Logger.SyslogFacility = getEnv("SYSLOG_FACILITY", "local1")
	cfg.Logger.SyslogHost = getEnv("SYSLOG_HOST", "")
	cfg.Logger.SyslogPort = getEnv("SYSLOG_PORT", "514")
	cfg.Logger.SyslogTag = getEnv("SYSLOG_TAG", "open311api")

	cfg.Sentry.DSN = getEnv("SENTRY_DSN", "")
	cfg.Sentry.Environment = getEnv("SENTRY_ENVIRONMENT", "development")
	cfg.Sentry.EnableTracing = getEnvBool("SENTRY_ENABLE_TRACING", false)
	cfg.Sentry.TracesSampleRate = getEnvFloat("SENTRY_TRACES_SAMPLE_RATE", 0.0)
	cfg.Sentry.SendDefaultPII = getEnvBool("SENTRY_SEND_DEFAULT_PII", false)

	if cfg.MongoDB.URI == "" {
		return nil, fmt.Errorf("MONGODB_URI is required")
	}

	return cfg, nil
}

// LoadDotEnv loads KEY=VALUE pairs from the given file into the process
// environment. Existing environment variables are never overwritten, so real
// env vars take precedence over the file. A missing file is not an error.
func LoadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getEnvFloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}
