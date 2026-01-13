package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type ModelProvider string

const (
	ProviderOpenAI   ModelProvider = "openai"
	ProviderZhipu    ModelProvider = "zhipu"
	ProviderQwen     ModelProvider = "qwen"
	ProviderDoubao   ModelProvider = "doubao"
	ProviderDeepSeek ModelProvider = "deepseek"
)

type ModelConfig struct {
	Provider  ModelProvider `json:"provider"`
	APIKey    string        `json:"api_key"`
	BaseURL   string        `json:"base_url"`
	ModelName string        `json:"model_name"`
}

type FileConfig struct {
	CurrentProvider ModelProvider          `json:"current_provider"`
	Models          map[string]ModelConfig `json:"models"`
}

func NewModel(ctx context.Context, cfg ModelConfig) (llms.Model, error) {
	switch cfg.Provider {
	case ProviderOpenAI:
		return openai.New(
			openai.WithToken(cfg.APIKey),
			openai.WithModel(cfg.ModelName),
			openai.WithBaseURL(cfg.BaseURL),
		)
	case ProviderZhipu:
		// Zhipu AI compatible with OpenAI
		// BaseURL: https://open.bigmodel.cn/api/paas/v4/
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://open.bigmodel.cn/api/paas/v4/"
		}
		return openai.New(
			openai.WithToken(cfg.APIKey),
			openai.WithModel(cfg.ModelName),
			openai.WithBaseURL(baseURL),
		)
	case ProviderQwen:
		// Qwen (DashScope) compatible with OpenAI
		// BaseURL: https://dashscope.aliyuncs.com/compatible-mode/v1
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		}
		return openai.New(
			openai.WithToken(cfg.APIKey),
			openai.WithModel(cfg.ModelName),
			openai.WithBaseURL(baseURL),
		)
	case ProviderDoubao:
		// Doubao (Volcengine) compatible with OpenAI
		// BaseURL: https://ark.cn-beijing.volces.com/api/v3
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://ark.cn-beijing.volces.com/api/v3"
		}
		return openai.New(
			openai.WithToken(cfg.APIKey),
			openai.WithModel(cfg.ModelName),
			openai.WithBaseURL(baseURL),
		)
	case ProviderDeepSeek:
		// DeepSeek compatible with OpenAI
		// BaseURL: https://api.deepseek.com
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://api.deepseek.com"
		}
		return openai.New(
			openai.WithToken(cfg.APIKey),
			openai.WithModel(cfg.ModelName),
			openai.WithBaseURL(baseURL),
		)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

// LoadFileConfig loads the full configuration from the json file
func LoadFileConfig(filePath string) (*FileConfig, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var fileConfig FileConfig
	if err := json.Unmarshal(file, &fileConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return &fileConfig, nil
}

// GetConfigFromFile helper to read config from json file and return the config for the current provider
func GetConfigFromFile(filePath string) (ModelConfig, error) {
	fileConfig, err := LoadFileConfig(filePath)
	if err != nil {
		return ModelConfig{}, err
	}

	if fileConfig.CurrentProvider == "" {
		return ModelConfig{}, fmt.Errorf("current_provider not set in config file")
	}

	modelConfig, ok := fileConfig.Models[string(fileConfig.CurrentProvider)]
	if !ok {
		return ModelConfig{}, fmt.Errorf("config for provider %s not found", fileConfig.CurrentProvider)
	}

	// Ensure provider is set correctly in the model config
	modelConfig.Provider = fileConfig.CurrentProvider
	return modelConfig, nil
}

// GetConfigFromEnv helper to read config from environment variables
// Env vars:
// LLM_PROVIDER: openai, zhipu, qwen, doubao, fake (default: fake)
// LLM_API_KEY: your api key
// LLM_MODEL: model name (e.g., gpt-4, glm-4, qwen-turbo, doubao-pro-4k)
// LLM_BASE_URL: optional custom base url
func GetConfigFromEnv() (ModelConfig, error) {
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		return ModelConfig{}, fmt.Errorf("LLM_PROVIDER not set")
	}

	return ModelConfig{
		Provider:  ModelProvider(provider),
		APIKey:    os.Getenv("LLM_API_KEY"),
		BaseURL:   os.Getenv("LLM_BASE_URL"),
		ModelName: os.Getenv("LLM_MODEL"),
	}, nil
}
