package rpc

import (
	"os"
	"time"

	"github.com/cloudwego/kitex/client"
)

// getClientOptions returns the common client options for RPC clients
func getClientOptions(envVarName, defaultAddr string) []client.Option {
	addr := os.Getenv(envVarName)
	if addr == "" {
		addr = defaultAddr
	}

	return []client.Option{
		client.WithHostPorts(addr),
		client.WithConnectTimeout(3 * time.Second),
		client.WithRPCTimeout(60 * time.Second),
	}
}
