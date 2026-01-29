package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Neo4j  *Neo4jConfig  `json:"neo4j"`
	Server *ServerConfig `json:"server"`
}

type Neo4jConfig struct {
	URI      string `json:"uri"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ServerConfig struct {
	Addr string `json:"addr"`
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
