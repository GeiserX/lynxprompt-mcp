package config

import (
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("LYNXPROMPT_URL", "")
	t.Setenv("LYNXPROMPT_TOKEN", "")
	cfg := LoadConfig()
	if cfg.BaseURL != "https://lynxprompt.geiser.cloud" {
		t.Errorf("BaseURL default: got %q, want %q", cfg.BaseURL, "https://lynxprompt.geiser.cloud")
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

func TestLoadConfig_LegacyURL(t *testing.T) {
	for in, want := range map[string]string{
		"https://lynxprompt.com":            "https://lynxprompt.geiser.cloud",
		"https://lynxprompt.com/":           "https://lynxprompt.geiser.cloud",
		"https://lynxprompt.geiser.cloud":   "https://lynxprompt.geiser.cloud",
		"https://lp.example.org":            "https://lp.example.org",
		"https://lynxprompt.com/custompath": "https://lynxprompt.com/custompath",
		"http://lynxprompt.com":             "http://lynxprompt.com",
	} {
		t.Setenv("LYNXPROMPT_URL", in)
		if got := LoadConfig().BaseURL; got != want {
			t.Errorf("LYNXPROMPT_URL=%q: BaseURL = %q, want %q", in, got, want)
		}
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

func TestLoadConfig_ReadOnly(t *testing.T) {
	for val, want := range map[string]bool{"": false, "0": false, "false": false, "1": true, "true": true, "TRUE": true, "yes": true} {
		t.Setenv("LYNXPROMPT_READ_ONLY", val)
		if got := LoadConfig().ReadOnly; got != want {
			t.Errorf("LYNXPROMPT_READ_ONLY=%q: ReadOnly = %v, want %v", val, got, want)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		url       string
		allowHTTP bool
		ok        bool
	}{
		{"https://lynxprompt.geiser.cloud", false, true},
		{"https://lp.example.org/", false, true},
		{"http://localhost:3000", false, true},
		{"http://127.0.0.1:3000", false, true},
		{"http://[::1]:3000", false, true},
		{"http://lp.example.org", false, false},
		{"http://lp.example.org", true, true},
		{"ftp://lp.example.org", true, false},
		{"lynxprompt.com", false, false},
		{"", false, false},
	}
	for _, tc := range tests {
		err := LynxPromptConfig{BaseURL: tc.url, AllowHTTP: tc.allowHTTP}.Validate()
		if (err == nil) != tc.ok {
			t.Errorf("Validate(%q, allowHTTP=%v) = %v, want ok=%v", tc.url, tc.allowHTTP, err, tc.ok)
		}
	}
}
