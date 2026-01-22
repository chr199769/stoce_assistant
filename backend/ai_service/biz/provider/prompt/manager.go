package prompt

import (
	"context"
	"log"
	"stock_assistant/backend/ai_service/biz/provider/langfuse"
)

type Manager struct {
	lf *langfuse.LangfuseManager
}

var globalManager *Manager

func Init(lf *langfuse.LangfuseManager) {
	globalManager = &Manager{lf: lf}
}

func GetManager() *Manager {
	return globalManager
}

// GetPrompt tries to fetch prompt from Langfuse, falls back to local default
func (m *Manager) GetPrompt(ctx context.Context, key string) string {
	// 1. Try Langfuse
	if m.lf != nil {
		if remotePrompt, err := m.lf.GetPrompt(ctx, key, nil); err == nil && remotePrompt != "" {
			log.Printf("Using remote prompt for %s (Source: Langfuse)", key)
			return remotePrompt
		} else {
			// Don't log too loudly on every request if fetch fails, maybe debug level
			log.Printf("Using local prompt for %s (Remote fetch failed: %v)", key, err)
		}
	}

	// 2. Fallback to local
	if p, ok := DefaultPrompts[key]; ok {
		return p
	}

	log.Printf("Warning: No prompt found for key %s", key)
	return ""
}
