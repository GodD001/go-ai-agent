package LLM_MCP_RAG

import (
	"context"
	"fmt"
	"testing"

	"github.com/joho/godotenv"

	"github.com/openai/openai-go/v3"
)

func TestChatGPTAI(t *testing.T) {
	godotenv.Load()
	ctx := context.Background()
	modelName := openai.ChatModelGPT3_5Turbo
	llm := NewChatOpenAI(ctx, modelName)
	prompt := "hello"
	result, toolCall := llm.Chat(prompt)
	if len(toolCall) != 0 {
		fmt.Println("toolCall:", toolCall)
	}
	fmt.Println("result:", result)
}
