package config_test

import (
	"strings"
	"testing"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/config"
)

func TestConfigEnvironmentNames(t *testing.T) {
	cfg, err := config.FromMap(map[string]string{
		"POSTGRES_SERVER": "db", "POSTGRES_USER": "u", "POSTGRES_PASSWORD": "p@ss/word", "POSTGRES_DB": "app",
		"BACKEND_CORS_ORIGINS": "http://a.test, http://b.test/", "FRONTEND_HOST": "http://f.test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppPort != 8000 || cfg.APIV1Str != "/api/v1" || cfg.AccessTokenExpireMinutes != 11520 || cfg.RedisPort != 6379 {
		t.Fatalf("defaults = %+v", cfg)
	}
	if got := strings.Join(cfg.CORSOrigins(), ","); got != "http://a.test,http://b.test,http://f.test" {
		t.Fatalf("cors = %s", got)
	}
	if got := cfg.PostgresURL(); got != "postgres://u:p%40ss%2Fword@db:5432/app" {
		t.Fatalf("pg url = %s", got)
	}
	if _, err := config.FromMap(map[string]string{"ENVIRONMENT": "nope"}); err == nil {
		t.Fatal("bad ENVIRONMENT accepted")
	}
	if _, err := config.FromMap(map[string]string{"APP_PORT": "x"}); err == nil {
		t.Fatal("bad APP_PORT accepted")
	}
}

func TestConfigRefusesChangethisOutsideLocal(t *testing.T) {
	for _, env := range []string{"staging", "production"} {
		cfg, _ := config.FromMap(map[string]string{"ENVIRONMENT": env})
		if cfg.Validate() == nil {
			t.Fatalf("%s accepted the default SECRET_KEY", env)
		}
		cfg, _ = config.FromMap(map[string]string{"ENVIRONMENT": env, "SECRET_KEY": "real"})
		if cfg.Validate() == nil {
			t.Fatalf("%s accepted the default FIRST_SUPERUSER_PASSWORD", env)
		}
		cfg, _ = config.FromMap(map[string]string{"ENVIRONMENT": env, "SECRET_KEY": "real", "FIRST_SUPERUSER_PASSWORD": "real"})
		if err := cfg.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	cfg, _ := config.FromMap(map[string]string{})
	if cfg.Validate() != nil {
		t.Fatal("local must accept defaults")
	}
}

func TestEnvFileParsing(t *testing.T) {
	got := config.ParseEnvFile("# c\nA=1\nexport B=\"two words\"\nC='x' \nD=v # note\n\nbad line\n")
	if got["A"] != "1" || got["B"] != "two words" || got["C"] != "x" || got["D"] != "v" || len(got) != 4 {
		t.Fatalf("parsed = %v", got)
	}
}
