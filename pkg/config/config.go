package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the JWKS mock service
type Config struct {
	Server ServerConfig `yaml:"server"`
	JWT    JWTConfig    `yaml:"jwt"`
	Keys   KeysConfig   `yaml:"keys"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Issuer   string `yaml:"issuer"`
	Audience string `yaml:"audience"`
}

// KeysConfig holds key generation configuration
type KeysConfig struct {
	Count  int      `yaml:"count"`
	KeyIDs []string `yaml:"key_ids"`
}

// Load loads configuration from environment variables and optional config file
func Load(configFile string) (*Config, error) {
	// Default configuration
	config := &Config{
		Server: ServerConfig{
			Port: 3000,
			Host: "0.0.0.0",
		},
		JWT: JWTConfig{
			Issuer:   "http://localhost:3000",
			Audience: "dev-api",
		},
		Keys: KeysConfig{
			Count:  2,
			KeyIDs: []string{"key-1", "key-2"},
		},
	}

	// Load from config file if provided
	if configFile != "" {
		if err := loadFromFile(config, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	return config, nil
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(config *Config, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if host := os.Getenv("HOST"); host != "" {
		config.Server.Host = host
	}

	if issuer := os.Getenv("JWT_ISSUER"); issuer != "" {
		config.JWT.Issuer = issuer
	}

	if audience := os.Getenv("JWT_AUDIENCE"); audience != "" {
		config.JWT.Audience = audience
	}

	if keyIDs := os.Getenv("KEY_IDS"); keyIDs != "" {
		ids := strings.Split(keyIDs, ",")
		for i := range ids {
			ids[i] = strings.TrimSpace(ids[i])
		}
		config.Keys.KeyIDs = ids
		config.Keys.Count = len(ids)
	} else if keyCount := os.Getenv("KEY_COUNT"); keyCount != "" {
		if count, err := strconv.Atoi(keyCount); err == nil && count > 0 {
			config.Keys.Count = count
			// Generate generic key IDs based on count if no specific IDs provided
			if len(config.Keys.KeyIDs) == 0 || len(config.Keys.KeyIDs) != count {
				config.Keys.KeyIDs = make([]string, count)
				for i := 0; i < count; i++ {
					config.Keys.KeyIDs[i] = fmt.Sprintf("key-%d", i+1)
				}
			}
		}
	}
}
