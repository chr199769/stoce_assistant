package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"stock_assistant/backend/ai_service/biz/provider/llm/core"
	"stock_assistant/backend/common/langfuse"
)

type Config struct {
	LLMConfig *core.FileConfig
	Langfuse  *langfuse.LangfuseConfig
	Runtime   *RuntimeConfig `json:"runtime"`
	RPC       *RPCConfig     `json:"rpc"`
	Server    *ServerConfig  `json:"server"`
}

var globalConfig *Config

func Init() error {
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

	// 2. Unmarshal Langfuse + Runtime + RPC + Server
	var wrapper struct {
		Langfuse *langfuse.LangfuseConfig `json:"langfuse"`
		Runtime  *RuntimeConfig           `json:"runtime"`
		RPC      *RPCConfig               `json:"rpc"`
		Server   *ServerConfig            `json:"server"`
	}
	if err := json.Unmarshal(file, &wrapper); err != nil {
		fmt.Printf("Warning: failed to unmarshal langfuse config: %v\n", err)
	}
	cfg.Langfuse = wrapper.Langfuse
	cfg.Runtime = wrapper.Runtime
	cfg.RPC = wrapper.RPC
	cfg.Server = wrapper.Server

	globalConfig = cfg
	if cfg.Runtime != nil && cfg.Runtime.LLMMaxConcurrent > 0 {
		core.SetLLMMaxConcurrent(cfg.Runtime.LLMMaxConcurrent)
	}
	if cfg.Runtime != nil {
		retryCfg := core.DefaultRetryConfig()
		if cfg.Runtime.LLMRetryMaxAttempts > 0 {
			retryCfg.MaxAttempts = cfg.Runtime.LLMRetryMaxAttempts
		}
		if cfg.Runtime.LLMRetryBaseDelayMs > 0 {
			retryCfg.BaseDelay = time.Duration(cfg.Runtime.LLMRetryBaseDelayMs) * time.Millisecond
		}
		if cfg.Runtime.LLMRetryMaxDelayMs > 0 {
			retryCfg.MaxDelay = time.Duration(cfg.Runtime.LLMRetryMaxDelayMs) * time.Millisecond
		}
		core.SetRetryConfig(retryCfg)
		if cfg.Runtime.LLMMinIntervalMs > 0 {
			core.SetLLMMinInterval(time.Duration(cfg.Runtime.LLMMinIntervalMs) * time.Millisecond)
		}
	}
	return nil
}

func Get() *Config {
	return globalConfig
}

type RuntimeConfig struct {
	MultiAgentEnabled   bool `json:"multi_agent_enabled"`
	LLMMaxConcurrent    int  `json:"llm_max_concurrent"`
	LLMRetryMaxAttempts int  `json:"llm_retry_max_attempts"`
	LLMRetryBaseDelayMs int  `json:"llm_retry_base_delay_ms"`
	LLMRetryMaxDelayMs  int  `json:"llm_retry_max_delay_ms"`
	LLMMinIntervalMs    int  `json:"llm_min_interval_ms"`
}

type RPCConfig struct {
	StockServiceAddr   string `json:"stock_service_addr"`
	KnowledgeGraphAddr string `json:"knowledge_graph_addr"`
}

type ServerConfig struct {
	Addr string `json:"addr"`
}
