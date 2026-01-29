package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Server *ServerConfig `json:"server"`
	RPC    *RPCConfig    `json:"rpc"`
}

type ServerConfig struct {
	Addr string `json:"addr"`
}

type RPCConfig struct {
	AIServiceAddr       string `json:"ai_service_addr"`
	StockServiceAddr    string `json:"stock_service_addr"`
	KnowledgeGraphAddr  string `json:"knowledge_graph_addr"`
  AkShareServiceAddr  string `json:"akshare_service_addr"`
}

var globalConfig *Config

func Init() error {
	cfg := &Config{}
	cwd, _ := os.Getwd()
	configPath := "conf/prod.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join(cwd, "conf/prod.json")
	}
	file, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if err := json.Unmarshal(file, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	globalConfig = cfg
	return nil
}

func Get() *Config {
	return globalConfig
}
