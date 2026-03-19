package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TelegramToken string `yaml:"telegram_token"`
	OpenAIAPIKey  string `yaml:"openai_api_key"`
	OpenAIModel   string `yaml:"openai_model"`

	DatabaseURL  string  `yaml:"database_url"`
	AdminUserIDs []int64 `yaml:"admin_user_ids"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}

	if v := os.Getenv("TELEGRAM_TOKEN"); v != "" {
		c.TelegramToken = v
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		c.OpenAIAPIKey = v
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		c.DatabaseURL = v
	}

	if c.OpenAIModel == "" {
		c.OpenAIModel = "gpt-5.2"
	}

	if c.TelegramToken == "" || c.OpenAIAPIKey == "" || c.DatabaseURL == "" {
		return nil, fmt.Errorf("missing telegram_token or openai_api_key or database_url")
	}

	return &c, nil
}
