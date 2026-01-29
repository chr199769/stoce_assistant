package vision

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"stock_assistant/backend/ai_service/biz/provider/llm/core"
	"stock_assistant/backend/ai_service/biz/provider/prompt"
	ai "stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/common/langfuse"

	"github.com/tmc/langchaingo/llms"
)

type Provider struct {
	fileConfig *core.FileConfig
}

func NewProvider(fileConfig *core.FileConfig) *Provider {
	return &Provider{
		fileConfig: fileConfig,
	}
}

func (p *Provider) RecognizeImage(ctx context.Context, imageData []byte, modelName string) ([]*ai.RecognizedStock, error) {
	// 1. 确定模型配置
	var cfg core.ModelConfig
	if p.fileConfig != nil {
		// 在所有提供商中查找模型配置
		found := false
		if modelName != "" {
			for provider, c := range p.fileConfig.Models {
				if c.ModelName == modelName {
					cfg = c
					cfg.Provider = core.ModelProvider(provider)
					found = true
					break
				}
			}
		}

		if !found {
			var ok bool
			cfg, ok = p.fileConfig.Models[string(p.fileConfig.CurrentProvider)]
			if ok {
				cfg.Provider = p.fileConfig.CurrentProvider
			} else {
				return nil, fmt.Errorf("未找到提供商且禁用了 fake 提供商")
			}
		}
	} else {
		return nil, fmt.Errorf("未找到配置且禁用了 fake 提供商")
	}

	log.Printf("使用 LLM 提供商进行图像识别: %s, 模型: %s", cfg.Provider, cfg.ModelName)

	if lf := langfuse.GetLangfuse(); lf != nil {
		traceID := lf.CreateTrace(ctx, "ImageRecognition", map[string]interface{}{
			"model": cfg.ModelName,
		})
		ctx = langfuse.WithTraceID(ctx, traceID)
	}

	// 2. 创建 LLM
	llmClient, err := core.NewModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 llm 失败: %w", err)
	}

	// 3. 准备图像
	// 检测内容类型
	mimeType := http.DetectContentType(imageData)
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	// 4. 创建消息
	promptStr := prompt.GetManager().GetPrompt(ctx, prompt.ImageRecognitionMaster)
	if promptStr == "" {
		return nil, fmt.Errorf("获取图像识别提示词失败")
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: promptStr},
				llms.ImageURLContent{URL: imageURL},
			},
		},
	}

	// 5. 生成内容
	log.Printf("[DEBUG] RecognizeImage Prompt:\n%s", promptStr)
	ctx = core.WithModelName(ctx, cfg.ModelName)
	ctx = core.WithGenerationName(ctx, "LLM-ImageRecognition")
	resp, err := core.GenerateContentWithRetry(ctx, llmClient, messages)
	if err != nil {
		return nil, fmt.Errorf("生成内容失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("未生成内容")
	}

	content := resp.Choices[0].Content
	log.Printf("LLM 原始响应: %s", content)

	// 6. 解析 JSON
	// 清理可能的 markdown 代码块
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var stocks []*ai.RecognizedStock
	// 使用临时结构体解析以处理可能的字段不匹配
	var tempStocks []struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal([]byte(content), &tempStocks); err != nil {
		// 如果 JSON 解析失败，尝试使用正则
		log.Printf("JSON 解析失败: %v, 尝试使用正则", err)
		// 简单正则匹配 sh/sz + 6位数字 或 纯6位数字
		re := regexp.MustCompile(`((sh|sz)\d{6})|(\d{6})`)
		matches := re.FindAllString(content, -1)

		uniqueCodes := make(map[string]bool)

		for _, match := range matches {
			code := match
			// 如果缺少前缀则修复
			if len(code) == 6 && !strings.HasPrefix(code, "sh") && !strings.HasPrefix(code, "sz") {
				if strings.HasPrefix(code, "6") {
					code = "sh" + code
				} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
					code = "sz" + code
				}
				// 忽略其他情况
			}

			// 避免重复并确保格式有效
			if !uniqueCodes[code] && (strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz")) {
				uniqueCodes[code] = true
				stocks = append(stocks, &ai.RecognizedStock{
					Code: code,
					Name: "未知", // 无法通过正则可靠提取名称
				})
			}
		}

		if len(stocks) == 0 {
			// 如果正则也未找到任何内容，返回空列表而不是错误
			log.Printf("从图像解析股票信息失败 (JSON 和正则均失败): %v", err)
			return []*ai.RecognizedStock{}, nil
		}
	} else {
		for _, s := range tempStocks {
			code := s.Code
			// 如果缺少前缀则修复
			if len(code) == 6 && !strings.HasPrefix(code, "sh") && !strings.HasPrefix(code, "sz") {
				if strings.HasPrefix(code, "6") {
					code = "sh" + code
				} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
					code = "sz" + code
				}
				// 忽略其他情况
			}

			stocks = append(stocks, &ai.RecognizedStock{
				Code: code,
				Name: s.Name,
			})
		}
	}

	return stocks, nil
}
