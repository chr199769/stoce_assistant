package llm

import (
	"context"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

type FakeModel struct{}

func NewFakeModel() *FakeModel {
	return &FakeModel{}
}

func (m *FakeModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	resp, err := m.GenerateContent(ctx, []llms.MessageContent{
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: prompt}}},
	}, options...)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Content, nil
}

func (m *FakeModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	lastMsg := messages[len(messages)-1]
	text := ""
	for _, part := range lastMsg.Parts {
		if t, ok := part.(llms.TextContent); ok {
			text += t.Text
		}
	}

	response := ""
	
	// Try to find the stock code from the prompt
	re := regexp.MustCompile(`Predict the trend for (\w+)`)
	matches := re.FindStringSubmatch(text)
	code := "sh600519"
	if len(matches) > 1 {
		code = matches[1]
	}

	hasPrice := strings.Contains(text, "Price:") && strings.Contains(text, "Volume:")
	hasNews := strings.Contains(text, "Recent financial report") || strings.Contains(text, "Industry sector")

	if hasPrice && hasNews {
		response = `Thought: I have gathered the stock price and the news. The price is stable and the news is positive. I can now make a prediction.
Final Answer: Based on the current price and positive news, the stock ` + code + ` is predicted to rise in the coming days. Confidence is high.`
	} else if hasPrice {
		response = `Thought: I have the stock price. Now I need to check the recent news to understand the market sentiment.
Action: StockNews
Action Input: ` + code
	} else {
		response = `Thought: I need to check the stock price for ` + code + `.
Action: StockPrice
Action Input: ` + code
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{Content: response},
		},
	}, nil
}
