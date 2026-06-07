package config

import (
	"encoding/json"
	"os"
)

// MongoDBConfig holds the MongoDB connection settings.
type MongoDBConfig struct {
	URI              string `json:"uri"`
	Database         string `json:"database"`
	Collection       string `json:"collection"`
	ConnectTimeout   int    `json:"connectTimeoutSeconds"`
	OperationTimeout int    `json:"operationTimeoutSeconds"`
	// TLSCertificateKeyFile is the path to the PEM file containing the client
	// certificate and private key, used for MONGODB-X509 authentication.
	TLSCertificateKeyFile string `json:"tlsCertificateKeyFile"`
	// TLSCAFile is an optional path to a CA bundle. Leave empty to use the
	// system root CAs (sufficient for MongoDB Atlas).
	TLSCAFile string `json:"tlsCAFile"`
}

// Config represents the application configuration
type Config struct {
	Server struct {
		Port                   int `json:"port"`
		ReadTimeoutSeconds     int `json:"readTimeoutSeconds"`
		WriteTimeoutSeconds    int `json:"writeTimeoutSeconds"`
		IdleTimeoutSeconds     int `json:"idleTimeoutSeconds"`
		ShutdownTimeoutSeconds int `json:"shutdownTimeoutSeconds"`
	} `json:"server"`
	Logger struct {
		Level          string `json:"level"`
		Format         string `json:"format"`
		ApacheLogPath  string `json:"apacheLogPath"`
		ToSyslog       bool   `json:"toSyslog"`
		SyslogFacility string `json:"syslogFacility"`
		SyslogHost     string `json:"syslogHost"`
		SyslogPort     string `json:"syslogPort"`
		SyslogTag      string `json:"syslogTag"`
	} `json:"logger"`
	MongoDB MongoDBConfig `json:"mongodb"`
	Sentry  struct {
		DSN              string  `json:"dsn"`
		EnableTracing    bool    `json:"enableTracing"`
		TracesSampleRate float64 `json:"tracesSampleRate"`
		SendDefaultPII   bool    `json:"sendDefaultPII"`
	} `json:"sentry"`
}

// Load loads the configuration from the specified file
func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
