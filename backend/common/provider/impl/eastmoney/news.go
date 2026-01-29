package eastmoneyimpl

import (
	"context"
	"time"

	em "stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/common/provider"
)

type EMClientNews interface {
	GetStockNews(ctx context.Context, code string) ([]*em.StockNewsItem, error)
}

type News struct {
	client EMClientNews
}

func NewNews(c EMClientNews) *News {
	return &News{client: c}
}

func (n *News) GetStockNews(ctx context.Context, code string, limit int) ([]*provider.NewsItem, error) {
	items, err := n.client.GetStockNews(ctx, code)
	if err != nil {
		return nil, err
	}
	var out []*provider.NewsItem
	for _, it := range items {
		t, _ := time.Parse("2006-01-02 15:04:05", it.ShowTime)
		out = append(out, &provider.NewsItem{
			Title:       it.Title,
			Summary:     it.Summary,
			Source:      "EastMoney",
			URL:         it.Url,
			PublishedAt: t,
		})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (n *News) GetMarketNews(ctx context.Context, limit int) ([]*provider.NewsItem, error) {
	return []*provider.NewsItem{}, nil
}

func (n *News) GetLiveNews(ctx context.Context, limit int) ([]*provider.LiveNewsItem, error) {
	return []*provider.LiveNewsItem{}, nil
}
