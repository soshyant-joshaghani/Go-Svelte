// Package config reads the FoxG environment (names match Fast's settings).
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	APIV1Str                 string
	SecretKey                string
	AccessTokenExpireMinutes int
	FrontendHost             string
	Environment              string
	BackendCORSOrigins       []string
	ProjectName              string
	PostgresServer           string
	PostgresPort             int
	PostgresUser             string
	PostgresPassword         string
	PostgresDB               string
	RedisHost                string
	RedisPort                int
	RedisDB                  int
	RedisPassword            string
	FirstSuperuser           string
	FirstSuperuserPassword   string
	AppHost                  string
	AppPort                  int
	BcryptCost               int
	MigrationsDir            string
}

// Load reads the process environment, then .env (looked up in ., .., ../..).
// Real, non-empty environment variables win over the file.
func Load() (Config, error) {
	return FromMap(loadEnv())
}

// FromMap builds a Config from a flat map. Empty values count as unset.
func FromMap(m map[string]string) (Config, error) {
	text := func(key, def string) string {
		if v := m[key]; v != "" {
			return v
		}
		return def
	}
	var firstErr error
	num := func(key string, def int) int {
		v := m[key]
		if v == "" {
			return def
		}
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("invalid value for %s: %q", key, v)
		}
		return n
	}
	cfg := Config{
		APIV1Str:                 text("API_V1_STR", "/api/v1"),
		SecretKey:                text("SECRET_KEY", "changethis"),
		AccessTokenExpireMinutes: num("ACCESS_TOKEN_EXPIRE_MINUTES", 60*24*8),
		FrontendHost:             text("FRONTEND_HOST", "http://dashboard.localhost"),
		Environment:              text("ENVIRONMENT", "local"),
		BackendCORSOrigins:       parseCORS(m["BACKEND_CORS_ORIGINS"]),
		ProjectName:              text("PROJECT_NAME", "go-svelte"),
		PostgresServer:           text("POSTGRES_SERVER", "localhost"),
		PostgresPort:             num("POSTGRES_PORT", 5432),
		PostgresUser:             text("POSTGRES_USER", "postgres"),
		PostgresPassword:         text("POSTGRES_PASSWORD", ""),
		PostgresDB:               text("POSTGRES_DB", ""),
		RedisHost:                text("REDIS_HOST", "localhost"),
		RedisPort:                num("REDIS_PORT", 6379),
		RedisDB:                  num("REDIS_DB", 0),
		RedisPassword:            text("REDIS_PASSWORD", ""),
		FirstSuperuser:           text("FIRST_SUPERUSER", "admin@example.com"),
		FirstSuperuserPassword:   text("FIRST_SUPERUSER_PASSWORD", "changethis"),
		AppHost:                  text("APP_HOST", "0.0.0.0"),
		AppPort:                  num("APP_PORT", 8000),
		BcryptCost:               num("BCRYPT_COST", 12),
		MigrationsDir:            m["MIGRATIONS_DIR"],
	}
	if firstErr != nil {
		return Config{}, firstErr
	}
	switch cfg.Environment {
	case "local", "staging", "production":
	default:
		return Config{}, fmt.Errorf("invalid value for ENVIRONMENT: %q", cfg.Environment)
	}
	return cfg, nil
}

// ForTests is a local config with a cheap bcrypt cost.
func ForTests() Config {
	cfg, err := FromMap(map[string]string{
		"ENVIRONMENT":              "local",
		"SECRET_KEY":               "test-secret-key",
		"FIRST_SUPERUSER":          "admin@example.com",
		"FIRST_SUPERUSER_PASSWORD": "adminpass123",
		"BCRYPT_COST":              "4",
	})
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c Config) IsLocal() bool { return c.Environment == "local" }

// Validate refuses the placeholder secrets outside ENVIRONMENT=local.
func (c Config) Validate() error {
	if c.IsLocal() {
		return nil
	}
	if c.SecretKey == "changethis" {
		return fmt.Errorf("SECRET_KEY must be set outside ENVIRONMENT=local")
	}
	if c.FirstSuperuserPassword == "changethis" {
		return fmt.Errorf("FIRST_SUPERUSER_PASSWORD must be set outside ENVIRONMENT=local")
	}
	return nil
}

// CORSOrigins is BACKEND_CORS_ORIGINS plus FRONTEND_HOST.
func (c Config) CORSOrigins() []string {
	out := append([]string(nil), c.BackendCORSOrigins...)
	host := strings.TrimRight(strings.TrimSpace(c.FrontendHost), "/")
	if host == "" {
		return out
	}
	for _, o := range out {
		if o == host {
			return out
		}
	}
	return append(out, host)
}

// PostgresURL is the connection string for pgx.
func (c Config) PostgresURL() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.PostgresUser, c.PostgresPassword),
		Host:   fmt.Sprintf("%s:%d", c.PostgresServer, c.PostgresPort),
		Path:   "/" + c.PostgresDB,
	}
	return u.String()
}

func parseCORS(raw string) []string {
	raw = strings.TrimSpace(raw)
	var items []string
	if strings.HasPrefix(raw, "[") {
		raw = strings.Trim(raw, "[]")
		for _, part := range strings.Split(raw, ",") {
			items = append(items, strings.Trim(strings.TrimSpace(part), "\"'"))
		}
	} else {
		items = strings.Split(raw, ",")
	}
	var out []string
	for _, item := range items {
		item = strings.TrimRight(strings.TrimSpace(item), "/")
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func loadEnv() map[string]string {
	m := map[string]string{}
	for _, dir := range []string{".", "..", filepath.Join("..", "..")} {
		data, err := os.ReadFile(filepath.Join(dir, ".env"))
		if err != nil {
			continue
		}
		for k, v := range ParseEnvFile(string(data)) {
			m[k] = v
		}
		break
	}
	for _, kv := range os.Environ() {
		if k, v, ok := strings.Cut(kv, "="); ok && v != "" {
			m[k] = v
		}
	}
	return m
}

// ParseEnvFile parses KEY=VALUE lines (comments, "export", simple quotes).
func ParseEnvFile(text string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, raw, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			continue
		}
		raw = strings.TrimSpace(raw)
		quoted := len(raw) >= 2 && ((raw[0] == '"' && raw[len(raw)-1] == '"') || (raw[0] == '\'' && raw[len(raw)-1] == '\''))
		if quoted {
			raw = raw[1 : len(raw)-1]
		} else if i := strings.Index(raw, " #"); i >= 0 {
			raw = strings.TrimSpace(raw[:i])
		}
		out[key] = raw
	}
	return out
}
