package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"stock_assistant/backend/stock_service/biz/provider/langfuse"

	"github.com/joho/godotenv"
)

type Config struct {
	Langfuse *langfuse.LangfuseConfig `json:"langfuse"`
}

var globalConfig *Config

func Init() error {
	// Load .env
	_ = godotenv.Load()

	cfg := &Config{}

	// Load LLM Config
	cwd, _ := os.Getwd()
	configPath := "conf/prod.json"
	// Try absolute path if relative fails
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join(cwd, "conf/prod.json")
	}
	file, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(file, &cfg); err != nil {
		fmt.Printf("Warning: failed to unmarshal langfuse config: %v\n", err)
	}

	globalConfig = cfg
	return nil
}

func Get() *Config {
	return globalConfig
}
