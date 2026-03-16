package LLM_MCP_RAG

import (
	"context"
	"fmt"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestChatGPTAI(t *testing.T) {
	ctx := context.Background()
	modelName := openai.ChatModelGPT3_5Turbo
	llm := newChatOpenAI(ctx, modelName)
	prompt := "What is the capital of France?"
	result, toolCall := llm.Chat(prompt)
	if len(toolCall) != 0 {
		fmt.Println("toolCall:", toolCall)
	}
	fmt.Println("result:", result)
}
