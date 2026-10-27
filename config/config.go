package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	if strings.ToLower(os.Getenv("TRANSPORT")) != "stdio" {
		_ = godotenv.Load()
	}
}

type LynxPromptConfig struct {
	BaseURL string
	Token   string
	// ReadOnly registers only the tools that read data (LYNXPROMPT_READ_ONLY).
	ReadOnly bool
	// AllowHTTP permits a plain http:// LYNXPROMPT_URL on a non-loopback host
	// (LYNXPROMPT_ALLOW_HTTP), which sends the token unencrypted.
	AllowHTTP bool
}

func LoadConfig() LynxPromptConfig {
	return LynxPromptConfig{
		BaseURL:   getEnv("LYNXPROMPT_URL", "https://lynxprompt.com"),
		Token:     getEnv("LYNXPROMPT_TOKEN", ""),
		ReadOnly:  envBool("LYNXPROMPT_READ_ONLY"),
		AllowHTTP: envBool("LYNXPROMPT_ALLOW_HTTP"),
	}
}

// Validate refuses a LYNXPROMPT_URL that is not an absolute http(s) URL, and
// a plain http:// URL to a remote host unless AllowHTTP is set, because the
// API token would travel in clear text.
func (c LynxPromptConfig) Validate() error {
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Host == "" {
		return fmt.Errorf("LYNXPROMPT_URL must be an absolute URL such as https://lynxprompt.com")
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if c.AllowHTTP || isLoopbackHost(u.Hostname()) {
			return nil
		}
		return fmt.Errorf("LYNXPROMPT_URL uses http:// on a non-loopback host, which would send LYNXPROMPT_TOKEN in clear text; use https:// or set LYNXPROMPT_ALLOW_HTTP=true")
	default:
		return fmt.Errorf("LYNXPROMPT_URL must start with https:// (got scheme %q)", u.Scheme)
	}
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envBool(k string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(k))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
