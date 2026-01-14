package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	AI       AIConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	DSN string
}

type AIConfig struct {
	Provider string // openai, ollama, etc.
	Model    string
	BaseURL  string
	APIKey   string
}

func Load() *Config {
	apiKey := getEnv("AI_API_KEY", "")
	// 同步设置 OPENAI_API_KEY，供 langchaingo 使用
	if apiKey != "" && os.Getenv("OPENAI_API_KEY") == "" {
		os.Setenv("OPENAI_API_KEY", apiKey)
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_DSN", "chat.db"),
		},
		AI: AIConfig{
			Provider: getEnv("AI_PROVIDER", "openai"),
			Model:    getEnv("AI_MODEL", "gpt-3.5-turbo"),
			BaseURL:  getEnv("AI_BASE_URL", ""),
			APIKey:   apiKey,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
