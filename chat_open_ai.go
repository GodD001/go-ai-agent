package LLM_MCP_RAG

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type ChatOpenAI struct {
	Ctx          context.Context
	ModelName    string
	SystemPrompt string
	RagContext   string
	Tools        []mcp.Tool
	Message      []openai.ChatCompletionMessageParamUnion
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
	if llm.SystemPrompt != "" {
		llm.Message = append(llm.Message, openai.SystemMessage(llm.SystemPrompt))
	}
	if llm.RagContext != "" {
		llm.Message = append(llm.Message, openai.UserMessage(llm.RagContext))
	}
	fmt.Println("Successfully created model:", llm.ModelName)
	return llm

}
func (c *ChatOpenAI) Chat(prompt string) (result string, toolCall []openai.ToolCallUnion) {
	if prompt != "" {
		c.Message = append(c.Message, openai.UserMessage(prompt))
	}
	toolParams := MCPTool2OpenAITool(c.Tools)
	stream := c.LLM.Chat.Completions.NewStreaming(c.Ctx, openai.ChatCompletionNewParams{
		Model:    c.ModelName,
		Messages: c.Message,
		Seed:     openai.Int(0),
		Tools:    toolParams,
	})
	result = ""
	finished := false

	acc := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)
		if content, ok := acc.JustFinishedContent(); ok {
			finished = true
			result = content
		}
		if tool, ok := acc.JustFinishedToolCall(); ok {
			fmt.Println("tool called", tool.Name)
			toolCall = append(toolCall, openai.ToolCallUnion{
				ID: tool.ID,
				Function: openai.FunctionToolCallFunction{
					Name:      tool.Name,
					Arguments: tool.Arguments,
				},
			})
		}
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if !finished {
				result += delta
			}
		}
	}
	if stream.Err() != nil {
		panic(stream.Err())
	}
	return result, toolCall
}
func MCPTool2OpenAITool(mcpTools []mcp.Tool) []openai.ChatCompletionToolUnionParam {
	openAITools := make([]openai.ChatCompletionToolUnionParam, 0, len(mcpTools))
	for _, tool := range mcpTools {
		params := openai.FunctionParameters{
			"type":       tool.InputSchema.Type,
			"properties": tool.InputSchema.Properties,
			"required":   tool.InputSchema.Required,
		}
		openAITools = append(openAITools, openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: shared.FunctionDefinitionParam{
					Name:        tool.Name,
					Description: openai.String(tool.Description),
					Parameters:  params,
				},
			},
		})
	}

	// Implementation for converting MCP tools to OpenAI tools
	return openAITools
}
