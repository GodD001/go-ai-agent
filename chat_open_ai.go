package LLM_MCP_RAG

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type ChatOpenAI struct {
	Ctx          context.Context
	ModelName    string
	SystemPrompt string
	RagContext   string
	Tools        []mcp.Tool
	LLM          openai.Client
}
type LLMOpition func(*ChatOpenAI)

func WithSystemPrompt(prompt string) LLMOpition {
	return func(c *ChatOpenAI) {
		c.SystemPrompt = prompt
	}
}
func WithRagContext(context string) LLMOpition {
	return func(c *ChatOpenAI) {
		c.RagContext = context
	}
}
func WithTools(tools []mcp.Tool) LLMOpition {
	return func(c *ChatOpenAI) {
		c.Tools = tools

	}
}

func newChatOpenAI(ctx context.Context, modelName string, opts ...LLMOpition) *ChatOpenAI {
	if modelName == "" {
		panic("modelName is empty")
	}
	var (
		apiKey  = os.Getenv("OPENAI_API")
		baseURL = os.Getenv("OPENAI_API_URL")
	)
	if apiKey == "" {
		panic("OPENAI_API not set")
	}
	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if baseURL != "" {
		options = append(options, option.WithBaseURL(baseURL))
	}
	cli := openai.NewClient(options...)
	llm := &ChatOpenAI{
		Ctx:       ctx,
		ModelName: modelName,
		LLM:       cli,
	}
	for _, opt := range opts {
		opt(llm)
	}
	return llm

}
