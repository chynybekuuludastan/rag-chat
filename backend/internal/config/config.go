package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	LLMProvider string
	OpenAI      OpenAIConfig
	Gemini      GeminiConfig
	JWT         JWTConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL string
}

type OpenAIConfig struct {
	APIKey         string
	ChatModel      string
	EmbeddingModel string
}

type GeminiConfig struct {
	APIKey         string
	ChatModel      string
	EmbeddingModel string
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		LLMProvider: getEnv("LLM_PROVIDER", "gemini"),
		OpenAI: OpenAIConfig{
			APIKey:         os.Getenv("OPENAI_API_KEY"),
			ChatModel:      getEnv("OPENAI_CHAT_MODEL", "gpt-4o-mini"),
			EmbeddingModel: getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
		},
		Gemini: GeminiConfig{
			APIKey:         os.Getenv("GEMINI_API_KEY"),
			ChatModel:      getEnv("GEMINI_CHAT_MODEL", "gemini-2.5-flash"),
			EmbeddingModel: getEnv("GEMINI_EMBEDDING_MODEL", "gemini-embedding-001"),
		},
		JWT: JWTConfig{
			Secret:          os.Getenv("JWT_SECRET"),
			AccessTokenTTL:  getDurationMinutes("JWT_ACCESS_TTL_MINUTES", 15),
			RefreshTokenTTL: getDurationMinutes("JWT_REFRESH_TTL_MINUTES", 10080), // 7 days
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	switch c.LLMProvider {
	case "gemini":
		if c.Gemini.APIKey == "" {
			return fmt.Errorf("GEMINI_API_KEY is required when LLM_PROVIDER=gemini")
		}
	case "openai":
		if c.OpenAI.APIKey == "" {
			return fmt.Errorf("OPENAI_API_KEY is required when LLM_PROVIDER=openai")
		}
	default:
		return fmt.Errorf("LLM_PROVIDER must be 'gemini' or 'openai', got %q", c.LLMProvider)
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationMinutes(key string, fallbackMinutes int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if minutes, err := strconv.Atoi(v); err == nil {
			return time.Duration(minutes) * time.Minute
		}
	}
	return time.Duration(fallbackMinutes) * time.Minute
}
