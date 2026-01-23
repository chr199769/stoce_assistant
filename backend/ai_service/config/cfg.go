package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"stock_assistant/backend/ai_service/biz/provider/langfuse"
	"stock_assistant/backend/ai_service/biz/provider/llm/core"

	"github.com/joho/godotenv"
)

type Config struct {
	LLMConfig *core.FileConfig
	Langfuse  *langfuse.LangfuseConfig
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

	// 1. Unmarshal LLM Config
	llmConfig := &core.FileConfig{}
	if err := json.Unmarshal(file, llmConfig); err != nil {
		return fmt.Errorf("failed to unmarshal llm config: %w", err)
	}
	cfg.LLMConfig = llmConfig

	// 2. Unmarshal Langfuse Config
	var wrapper struct {
		Langfuse *langfuse.LangfuseConfig `json:"langfuse"`
	}
	if err := json.Unmarshal(file, &wrapper); err != nil {
		fmt.Printf("Warning: failed to unmarshal langfuse config: %v\n", err)
	}
	cfg.Langfuse = wrapper.Langfuse

	globalConfig = cfg
	return nil
}

func Get() *Config {
	return globalConfig
}
