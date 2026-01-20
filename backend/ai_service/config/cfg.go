package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"stock_assistant/backend/ai_service/biz/provider/llm"

	"github.com/joho/godotenv"
)

type Config struct {
	LLMConfig *llm.FileConfig
	Langfuse  *llm.LangfuseConfig
}

var globalConfig *Config

func Init() error {
	// Load .env
	_ = godotenv.Load()

	cfg := &Config{}

	// Load LLM Config
	cwd, _ := os.Getwd()
	configPath := "conf/llm_config.json"
	// Try absolute path if relative fails
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join(cwd, "conf/llm_config.json")
	}
	file, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(file, cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	globalConfig = cfg
	return nil
}

func Get() *Config {
	return globalConfig
}
