package LLM_MCP_RAG

import (
	"context"
	"testing"
)

func TestMCPClient(t *testing.T) {
	ctx := context.Background()
	client := NewMCPClient(ctx, "uvx", nil, []string{"mcp-server-fetch"})
	err := client.Start()
	if err != nil {
		t.Fatalf("failed to start MCP client: %v", err)
	}
	err = client.SetTools()
	if err != nil {
		t.Fatalf("failed to set tools: %v", err)
	}
	tools := client.GetTools()
	if len(tools) == 0 {
		t.Fatalf("failed to get tools: %v", err)
	}
}
