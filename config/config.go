package config

import (
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
}

func LoadConfig() LynxPromptConfig {
	return LynxPromptConfig{
		BaseURL: getEnv("LYNXPROMPT_URL", "https://lynxprompt.com"),
		Token:   getEnv("LYNXPROMPT_TOKEN", ""),
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
