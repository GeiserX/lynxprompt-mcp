package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env in the working directory; ignore error if the file is absent.
	_ = godotenv.Load()
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
