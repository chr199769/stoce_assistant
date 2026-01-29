package eastmoneyimpl

import (
	"context"
	"time"

	em "stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/common/provider"
)

type EMClientNotice interface {
	GetStockNotices(ctx context.Context, code string, filterKeywords []string) ([]*em.NoticeItem, error)
}

type Notice struct {
	client EMClientNotice
}

func NewNotice(c EMClientNotice) *Notice {
	return &Notice{client: c}
}

func (n *Notice) GetStockNotices(ctx context.Context, code string, keyword string, limit int) ([]*provider.NoticeItem, error) {
	var kws []string
	if keyword != "" {
		kws = []string{keyword}
	}
	items, err := n.client.GetStockNotices(ctx, code, kws)
	if err != nil {
		return nil, err
	}
	var out []*provider.NoticeItem
	for _, it := range items {
		t, _ := time.Parse("2006-01-02", it.Date)
		out = append(out, &provider.NoticeItem{
			Title:       it.Title,
			Category:    "",
			URL:         it.Url,
			PublishedAt: t,
			Keywords:    kws,
		})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}
