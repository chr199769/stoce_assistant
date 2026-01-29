package prompt

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"stock_assistant/backend/common/langfuse"
)

type Manager struct {
	lf       *langfuse.LangfuseManager
	baseDir  string
	localMap map[string]string
}

var globalManager *Manager

func Init(lf *langfuse.LangfuseManager) {
	globalManager = &Manager{
		lf:       lf,
		baseDir:  "prompts",
		localMap: map[string]string{},
	}
}

func GetManager() *Manager {
	return globalManager
}

// GetPrompt tries to fetch prompt from Langfuse, falls back to local default
func (m *Manager) GetPrompt(ctx context.Context, key string) string {
	// 1. Try Langfuse
	// if m.lf != nil {
	// 	if remotePrompt, err := m.lf.GetPrompt(ctx, key, nil); err == nil && remotePrompt != "" {
	// 		log.Printf("Using remote prompt for %s (Source: Langfuse)", key)
	// 		return remotePrompt
	// 	} else {
	// 		// Don't log too loudly on every request if fetch fails, maybe debug level
	// 		log.Printf("Using local prompt for %s (Remote fetch failed: %v)", key, err)
	// 	}
	// }

	if prompt, ok := m.localMap[key]; ok && prompt != "" {
		return prompt
	}

	if prompt := m.loadPromptFromFile(key); prompt != "" {
		m.localMap[key] = prompt
		return prompt
	}

	log.Printf("Warning: No prompt found for key %s", key)
	return ""
}

func (m *Manager) loadPromptFromFile(key string) string {
	if m == nil {
		return ""
	}
	filename := key + ".prompt"
	path := filepath.Join(m.baseDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return ""
		}
		path = filepath.Join(cwd, m.baseDir, filename)
		content, err = os.ReadFile(path)
		if err != nil {
			return ""
		}
	}
	return strings.TrimSpace(string(content))
}
