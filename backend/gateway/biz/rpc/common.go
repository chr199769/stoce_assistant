package rpc

import (
	"time"

	"github.com/cloudwego/kitex/client"
)

// getClientOptions returns the common client options for RPC clients
func getClientOptions(addr string) []client.Option {
	return []client.Option{
		client.WithHostPorts(addr),
		client.WithConnectTimeout(3 * time.Second),
		client.WithRPCTimeout(60 * time.Second),
	}
}
