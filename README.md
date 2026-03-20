# LLM MCP RAG

A Go library that connects OpenAI language models to [Model Context Protocol (MCP)](https://modelcontextprotocol.io) servers. It provides a three-layer architecture: an MCP client for tool discovery and execution, an LLM chat layer for streaming completions, and an agent loop that drives multi-step tool use until the model produces a final answer.

## Architecture

```
┌─────────────┐       prompt / tool results       ┌──────────────┐
│    Agent    │ ────────────────────────────────▶ │  ChatOpenAI  │
│  (agent.go) │ ◀──────────────────────────────── │(chat_open_ai)│
└──────┬──────┘       text + tool calls           └──────────────┘
       │
       │  CallTool(name, args)
       ▼
┌─────────────┐
│  MCPClient  │  ── stdio ──▶  MCP Server (e.g. mcp-server-fetch)
│(mcp_client) │
└─────────────┘
```

The agent loop:

1. Send user prompt to the LLM
2. If the model returns tool calls, execute each one via the MCP server
3. Append tool results back into the conversation history
4. Ask the model again — repeat until no tool calls remain or `maxSteps` is reached

## Prerequisites

- Go 1.21+
- An OpenAI API key
- [uvx](https://docs.astral.sh/uv/) for running MCP servers (used in tests with `mcp-server-fetch`)

## Environment Variables

| Variable         | Required | Description                                      |
|------------------|----------|--------------------------------------------------|
| `OPENAI_API`     | Yes      | OpenAI API key                                   |
| `OPENAI_API_URL` | No       | Custom base URL (e.g. for a compatible endpoint) |

Create a `.env` file in the project root (loaded automatically in tests via `godotenv`):

```
OPENAI_API=sk-...
OPENAI_API_URL=https://api.openai.com/v1   # optional
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    llmmcprag "llm_mcp_rag"
    openai "github.com/openai/openai-go/v3"
)

func main() {
    ctx := context.Background()

    // 1. Start an MCP server over stdio
    mcpClient := llmmcprag.NewMCPClient(ctx, "uvx", nil, []string{"mcp-server-fetch"})
    if err := mcpClient.Start(); err != nil {
        panic(err)
    }
    defer mcpClient.Close()

    // 2. Discover tools from the MCP server
    if err := mcpClient.SetTools(); err != nil {
        panic(err)
    }

    // 3. Create the LLM client and attach the discovered tools
    llm := llmmcprag.NewChatOpenAI(
        ctx,
        openai.ChatModelGPT4oMini,
        llmmcprag.WithSystemPrompt("You are a helpful assistant. Use tools when needed."),
        llmmcprag.WithTools(mcpClient.GetTools()),
    )

    // 4. Run the agent loop
    agent := llmmcprag.NewAgent(llm, mcpClient)
    result, err := agent.Run("Fetch https://httpbin.org/get and summarise the response.", llmmcprag.DefaultMaxSteps)
    if err != nil {
        panic(err)
    }
    fmt.Println(result)
}
```

## API

### MCPClient

Connects to an external MCP server over stdio.

```go
// Create and start a client
client := NewMCPClient(ctx, cmd string, env []string, args []string)
err := client.Start()

// Discover tools from the server
err = client.SetTools()
tools := client.GetTools() // []mcp.Tool

// Call a specific tool
result, err := client.CallTool(toolName string, args any) // returns text output

// Shut down
err = client.Close()
```

### ChatOpenAI

Wraps the OpenAI streaming chat completions API.

```go
llm := NewChatOpenAI(ctx, modelName string, opts ...LLMOpition)

// Options
WithSystemPrompt(prompt string)   // prepend a system message
WithRagContext(context string)    // prepend a user message with retrieval context
WithTools(tools []mcp.Tool)       // expose MCP tools to the model

// Single-turn streaming call; returns the assistant text and any tool calls
result, toolCalls := llm.Chat(prompt string)
```

### Agent

Drives the multi-step tool-use loop.

```go
agent := NewAgent(llm *ChatOpenAI, mcp *MCPClient)

// Run until a final answer is produced or maxSteps is reached.
// Pass DefaultMaxSteps (10) or any positive integer.
result, err := agent.Run(prompt string, maxSteps int)
```

The loop handles:
- Argument JSON parse errors — reported back to the model as a tool error message
- MCP tool call failures — surfaced as a tool error message so the model can self-correct
- `IsError` responses from the MCP server — treated as errors

## Running Tests

```bash
# Copy and fill in your credentials first
cp .env.example .env   # or create .env manually

# Run all tests
go test ./...

# Run only the end-to-end agent test (requires uvx + network access)
go test -v -run TestAgent -timeout 120s
```

## Dependencies

| Package | Purpose |
|---------|---------|
| [`github.com/mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go) | MCP client and protocol types |
| [`github.com/openai/openai-go/v3`](https://github.com/openai/openai-go) | OpenAI API client with streaming support |
| [`github.com/joho/godotenv`](https://github.com/joho/godotenv) | `.env` file loading in tests |
