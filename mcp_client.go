package LLM_MCP_RAG

import (
	"context"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPClient struct {
	Ctx    context.Context
	Client *client.Client
	Cmd    string
	Tools  []mcp.Tool
	Args   []string
	Env    []string
}

func NewMCPClient(ctx context.Context, cmd string, env, args []string) *MCPClient {
	studioTransport := transport.NewStdio(cmd, env, args...)
	client := client.NewClient(studioTransport)
	return &MCPClient{
		Ctx:    ctx,
		Client: client,
		Cmd:    cmd,
		Args:   args,
		Env:    env,
	}
}
func (m *MCPClient) Start() error {
	err := m.Client.Start(m.Ctx)
	if err != nil {
		return err
	}
	mcpInitReq := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "LLM-MCP-RAG",
				Version: "0.1.0",
			},
		},
	}
	if _, err := m.Client.Initialize(m.Ctx, mcpInitReq); err != nil {
		return err
	}
	return nil
}
func (m *MCPClient) SetTools() error {
	toolsReq := mcp.ListToolsRequest{}
	tools, err := m.Client.ListTools(m.Ctx, toolsReq)
	if err != nil {
		return err
	}
	m.Tools = tools.Tools
	return nil

}
func (m *MCPClient) Close() error {
	return m.Client.Close()
}
func (m *MCPClient) CallTool(toolName string, args any) (string, error) {
	res, err := m.Client.CallTool(m.Ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	})
	if err != nil {
		return "", err
	}
	return mcp.GetTextFromContent(res.Result), nil

}
func (m *MCPClient) GetTools() []mcp.Tool {
	return m.Tools
}
