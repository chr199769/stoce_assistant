package llm

import (
	"context"
	"fmt"

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
	ProviderFake     ModelProvider = "fake"
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
		// 智谱 AI 兼容 OpenAI
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
		// 通义千问 (DashScope) 兼容 OpenAI
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
		// 豆包 (火山引擎) 兼容 OpenAI
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
		// DeepSeek 兼容 OpenAI
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
		return nil, fmt.Errorf("不支持的提供商: %s", cfg.Provider)
	}
}
