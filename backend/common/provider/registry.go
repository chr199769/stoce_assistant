package provider

import "fmt"

// ProviderBundle 统一封装各类客户端
type ProviderBundle struct {
	Market     MarketDataClient
	DragonTiger DragonTigerClient
	Financial  FinancialClient
	Sector     SectorClient
	News       NewsClient
	Notice     NoticeClient
	Popularity PopularityClient
	Sentiment  SentimentClient
}

// Factory 创建 ProviderBundle
type Factory func() (*ProviderBundle, error)

// Registry 注册中心
var registry = map[string]Factory{}

// Register 注册 provider
func Register(name string, factory Factory) {
	registry[name] = factory
}

// Resolve 获取 provider
func Resolve(name string) (*ProviderBundle, error) {
	if f, ok := registry[name]; ok {
		return f()
	}
	return nil, fmt.Errorf("provider 未注册: %s", name)
}
