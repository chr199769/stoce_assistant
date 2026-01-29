package core

import (
	"context"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"stock_assistant/backend/common/langfuse"

	"github.com/tmc/langchaingo/llms"
)

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   2 * time.Second,
		MaxDelay:    20 * time.Second,
	}
}

func GenerateContentWithRetry(ctx context.Context, model llms.Model, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if !acquireLLMSlot(ctx) {
		return nil, ctx.Err()
	}
	defer releaseLLMSlot()

	cfg := getRetryConfig()

	var lastErr error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if !waitForLLMInterval(ctx) {
			return nil, ctx.Err()
		}
		start := time.Now()
		resp, err := model.GenerateContent(ctx, messages, options...)
		end := time.Now()
		if err == nil {
			traceID := langfuse.TraceIDFromContext(ctx)
			if traceID != "" {
				lf := langfuse.GetLangfuse()
				if lf != nil {
					name := GenerationNameFromContext(ctx)
					if name == "" {
						name = "LLM-Generate"
					}
					modelName := ModelNameFromContext(ctx)
					metadata := GenerationMetadataFromContext(ctx)
					lfInput := buildLangfuseInput(messages)
					lfOutput := ""
					if resp != nil && len(resp.Choices) > 0 {
						lfOutput = resp.Choices[0].Content
					}
					lf.CreateGeneration(ctx, traceID, name, modelName, lfInput, lfOutput, metadata, start, end)
				}
			}
			return resp, nil
		}
		lastErr = err
		if !isRateLimitError(err) || attempt == cfg.MaxAttempts {
			return nil, err
		}
		wait := backoffDelay(cfg, attempt)
		if retryAfter := parseRetryAfter(err); retryAfter > 0 {
			wait = applyRetryAfter(wait, retryAfter, cfg)
		}
		log.Printf("触发限流重试，第%d次，等待%s，错误: %v", attempt, wait, err)
		if !sleepWithContext(ctx, wait) {
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}

var llmLimiterOnce sync.Once
var llmLimiter chan struct{}
var llmMaxConcurrent = 1
var llmMinInterval time.Duration
var llmIntervalMu sync.Mutex
var llmNextAllowed time.Time
var retryConfigMu sync.RWMutex
var retryConfig = DefaultRetryConfig()

func SetLLMMaxConcurrent(limit int) {
	if limit > 0 {
		llmMaxConcurrent = limit
	}
}

func SetLLMMinInterval(d time.Duration) {
	if d < 0 {
		d = 0
	}
	llmIntervalMu.Lock()
	llmMinInterval = d
	llmNextAllowed = time.Time{}
	llmIntervalMu.Unlock()
}

func SetRetryConfig(cfg RetryConfig) {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 1 * time.Second
	}
	if cfg.MaxDelay < 0 {
		cfg.MaxDelay = 0
	}
	retryConfigMu.Lock()
	retryConfig = cfg
	retryConfigMu.Unlock()
}

func getRetryConfig() RetryConfig {
	retryConfigMu.RLock()
	cfg := retryConfig
	retryConfigMu.RUnlock()
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 1 * time.Second
	}
	if cfg.MaxDelay < 0 {
		cfg.MaxDelay = 0
	}
	return cfg
}

func acquireLLMSlot(ctx context.Context) bool {
	llmLimiterOnce.Do(func() {
		llmLimiter = make(chan struct{}, llmMaxConcurrent)
	})
	select {
	case llmLimiter <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

func releaseLLMSlot() {
	select {
	case <-llmLimiter:
	default:
	}
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "rate_limit") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "throttle") ||
		strings.Contains(msg, "限流")
}

func backoffDelay(cfg RetryConfig, attempt int) time.Duration {
	delay := cfg.BaseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
	}
	if cfg.MaxDelay > 0 && delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	jitter := time.Duration(r.Int63n(int64(delay/3) + 1))
	delay += jitter
	if cfg.MaxDelay > 0 && delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}
	return delay
}

func sleepWithContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func waitForLLMInterval(ctx context.Context) bool {
	llmIntervalMu.Lock()
	minInterval := llmMinInterval
	now := time.Now()
	next := llmNextAllowed
	if minInterval <= 0 {
		llmIntervalMu.Unlock()
		return true
	}
	if next.Before(now) {
		next = now
	}
	wait := next.Sub(now)
	llmNextAllowed = next.Add(minInterval)
	llmIntervalMu.Unlock()
	if wait <= 0 {
		return true
	}
	return sleepWithContext(ctx, wait)
}

var retryAfterPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)retry[-\\s]?after[:=\\s]*(\\d+)(ms|s)?`),
	regexp.MustCompile(`(?i)等待\\s*(\\d+)\\s*(秒|毫秒|ms|s)`),
}

func parseRetryAfter(err error) time.Duration {
	if err == nil {
		return 0
	}
	msg := err.Error()
	for _, pattern := range retryAfterPatterns {
		matches := pattern.FindStringSubmatch(msg)
		if len(matches) < 2 {
			continue
		}
		value, convErr := strconv.Atoi(matches[1])
		if convErr != nil {
			continue
		}
		unit := "s"
		if len(matches) > 2 && matches[2] != "" {
			unit = strings.ToLower(matches[2])
		}
		switch unit {
		case "ms", "毫秒":
			return time.Duration(value) * time.Millisecond
		default:
			return time.Duration(value) * time.Second
		}
	}
	return 0
}

func applyRetryAfter(backoff time.Duration, retryAfter time.Duration, cfg RetryConfig) time.Duration {
	wait := backoff
	if retryAfter > wait {
		wait = retryAfter
	}
	if cfg.MaxDelay > 0 && wait > cfg.MaxDelay {
		return cfg.MaxDelay
	}
	return wait
}

func buildLangfuseInput(messages []llms.MessageContent) []map[string]interface{} {
	lfInput := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		var contentStr string
		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				contentStr += textPart.Text
			} else if imgPart, ok := part.(llms.ImageURLContent); ok {
				contentStr += "[Image:" + imgPart.URL + "]"
			}
		}
		roleStr := strings.ToLower(string(msg.Role))
		if roleStr == "human" {
			roleStr = "user"
		} else if roleStr == "ai" {
			roleStr = "assistant"
		}
		lfInput = append(lfInput, map[string]interface{}{
			"role":    roleStr,
			"content": contentStr,
		})
	}
	return lfInput
}
