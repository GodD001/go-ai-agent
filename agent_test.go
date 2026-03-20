package LLM_MCP_RAG

import (
	"context"
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	openai "github.com/openai/openai-go/v3"
)

// TestAgent runs a minimal end-to-end agent loop:
// MCP server (mcp-server-fetch) + OpenAI model.
// The agent is asked to fetch a URL so that it is forced to call the fetch tool.
func TestAgent(t *testing.T) {
	godotenv.Load()
	ctx := context.Background()

	mcpClient := NewMCPClient(ctx, "uvx", nil, []string{"mcp-server-fetch"})
	if err := mcpClient.Start(); err != nil {
		t.Fatalf("MCP start: %v", err)
	}
	defer mcpClient.Close()

	if err := mcpClient.SetTools(); err != nil {
		t.Fatalf("MCP SetTools: %v", err)
	}
	tools := mcpClient.GetTools()
	if len(tools) == 0 {
		t.Fatal("no MCP tools discovered")
	}
	fmt.Printf("discovered %d MCP tool(s)\n", len(tools))

	llm := NewChatOpenAI(
		ctx,
		openai.ChatModelGPT4oMini,
		WithSystemPrompt("You are a helpful assistant. Use tools when needed."),
		WithTools(tools),
	)

	agent := NewAgent(llm, mcpClient)
	result, err := agent.Run("请使用 fetch 工具获取 https://httpbin.org/get 的内容，并总结一下返回了什么。", DefaultMaxSteps)
	if err != nil {
		t.Fatalf("agent Run: %v", err)
	}
	if result == "" {
		t.Fatal("agent returned empty result")
	}
	fmt.Println("agent result:", result)
}
