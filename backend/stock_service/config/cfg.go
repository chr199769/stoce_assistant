package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"stock_assistant/backend/common/langfuse"
)

type Config struct {
	Langfuse *langfuse.LangfuseConfig `json:"langfuse"`
	Database *DatabaseConfig          `json:"database"`
	RPC      *RPCConfig               `json:"rpc"`
	AkShare  *AkShareConfig           `json:"akshare"`
}

var globalConfig *Config
var globalConfigPath string

type DatabaseConfig struct {
	MySQLDSN string `json:"mysql_dsn"`
}

type RPCConfig struct {
	KnowledgeGraphAddr string `json:"knowledge_graph_addr"`
	AIServiceAddr      string `json:"ai_service_addr"`
}

type AkShareConfig struct {
	PythonBin string             `json:"python_bin"`
	ServiceURL string            `json:"service_url"`
	News      *AkShareNewsConfig `json:"news"`
	Macro     *AkShareMacroConfig `json:"macro"`
}

type AkShareNewsConfig struct {
	Func   string `json:"func"`
	Symbol string `json:"symbol"`
	Limit  int    `json:"limit"`
}

type AkShareMacroConfig struct {
	Funcs  []string `json:"funcs"`
	Limit  int      `json:"limit"`
	Symbol string   `json:"symbol"`
}

func Init() error {
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
	globalConfigPath = configPath
	return nil
}

func Get() *Config {
	return globalConfig
}

func GetConfigPath() string {
	return globalConfigPath
}
