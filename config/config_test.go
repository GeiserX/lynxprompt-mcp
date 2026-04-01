package config

import (
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("LYNXPROMPT_URL", "")
	t.Setenv("LYNXPROMPT_TOKEN", "")
	cfg := LoadConfig()
	if cfg.BaseURL != "https://lynxprompt.com" {
		t.Errorf("BaseURL default: got %q, want %q", cfg.BaseURL, "https://lynxprompt.com")
	}
	if cfg.Token != "" {
		t.Errorf("Token default: got %q, want empty", cfg.Token)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	t.Setenv("LYNXPROMPT_URL", "http://localhost:9999")
	t.Setenv("LYNXPROMPT_TOKEN", "my-secret-token")
	cfg := LoadConfig()
	if cfg.BaseURL != "http://localhost:9999" {
		t.Errorf("BaseURL: got %q, want %q", cfg.BaseURL, "http://localhost:9999")
	}
	if cfg.Token != "my-secret-token" {
		t.Errorf("Token: got %q, want %q", cfg.Token, "my-secret-token")
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envVal   string
		fallback string
		want     string
	}{
		{
			name:     "returns env value when set",
			key:      "TEST_LYNX_VAR",
			envVal:   "custom",
			fallback: "default",
			want:     "custom",
		},
		{
			name:     "returns default when env empty",
			key:      "TEST_LYNX_EMPTY",
			envVal:   "",
			fallback: "fallback",
			want:     "fallback",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tc.key, tc.envVal)
			got := getEnv(tc.key, tc.fallback)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
