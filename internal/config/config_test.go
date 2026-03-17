package config

import (
	"testing"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
)

func TestApplyDerivedDefaultsSetsCookieSecureByEnvironment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		environment string
		wantSecure  bool
	}{
		{
			name:        "development defaults to insecure cookies",
			environment: "development",
			wantSecure:  false,
		},
		{
			name:        "production defaults to secure cookies",
			environment: "production",
			wantSecure:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			k := koanf.New(".")
			cfg := Config{}
			cfg.App.Environment = tt.environment

			applyDerivedDefaults(k, &cfg)

			if cfg.Auth.CookieSecure != tt.wantSecure {
				t.Fatalf("CookieSecure = %t, want %t", cfg.Auth.CookieSecure, tt.wantSecure)
			}
		})
	}
}

func TestApplyDerivedDefaultsPreservesExplicitCookieSecure(t *testing.T) {
	t.Parallel()

	k := koanf.New(".")
	if err := k.Load(confmap.Provider(map[string]interface{}{
		"auth.cookie_secure": false,
	}, "."), nil); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	cfg := Config{}
	cfg.App.Environment = "production"
	cfg.Auth.CookieSecure = false

	applyDerivedDefaults(k, &cfg)

	if cfg.Auth.CookieSecure {
		t.Fatal("CookieSecure was overridden despite explicit configuration")
	}
}

func TestBuildDatabaseURLEscapesReservedCharacters(t *testing.T) {
	t.Parallel()

	got := buildDatabaseURL("app-user", "p@ss:/word", "localhost", "5432", "go-web-server", "disable")
	want := "postgres://app-user:p%40ss%3A%2Fword@localhost:5432/go-web-server?sslmode=disable"

	if got != want {
		t.Fatalf("buildDatabaseURL() = %q, want %q", got, want)
	}
}
