package config

import (
	"encoding/json"
	"log/slog"
	"net/url"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Worker   WorkerConfig   `mapstructure:"worker"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Path DatabaseURL `mapstructure:"path"`
}

type WorkerConfig struct {
	Secret string `mapstructure:"secret"`
}

// DatabaseURL is a connection string that may embed credentials (e.g. a Turso
// libsql URL with an `authToken` query parameter). It redacts the credential
// when formatted, logged via slog, or JSON-marshaled, so the bare value
// reaches a writer only via an explicit `string(...)` conversion.
type DatabaseURL string

func (d DatabaseURL) LogValue() slog.Value {
	return slog.StringValue(d.redacted())
}

func (d DatabaseURL) String() string {
	return d.redacted()
}

func (d DatabaseURL) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.redacted())
}

func (d DatabaseURL) redacted() string {
	raw := string(d)
	u, err := url.Parse(raw)
	if err != nil {
		return "<unparseable>"
	}
	q := u.Query()
	if !q.Has("authToken") {
		return raw
	}
	q.Set("authToken", "REDACTED")
	u.RawQuery = q.Encode()
	return u.String()
}
