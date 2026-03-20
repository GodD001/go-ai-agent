package LLM_MCP_RAG

import (
	"encoding/json"
	"fmt"

	openai "github.com/openai/openai-go/v3"
)

const DefaultMaxSteps = 10

type Agent struct {
	LLM *ChatOpenAI
	MCP *MCPClient
}

func NewAgent(llm *ChatOpenAI, mcp *MCPClient) *Agent {
	return &Agent{LLM: llm, MCP: mcp}
}

// Run executes the agent loop: calls the LLM, executes any tool calls via MCP,
// feeds results back into the conversation, and repeats until the model produces
// a final text answer or maxSteps is reached.
func (a *Agent) Run(prompt string, maxSteps int) (string, error) {
	if maxSteps <= 0 {
		maxSteps = DefaultMaxSteps
	}

	result, toolCalls := a.LLM.Chat(prompt)

	for step := 0; step < maxSteps && len(toolCalls) > 0; step++ {
		// Build the assistant message that carries the tool_calls back into history.
		toolCallParams := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(toolCalls))
		for _, tc := range toolCalls {
			toolCallParams = append(toolCallParams, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID: tc.ID,
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				},
			})
		}
		a.LLM.Message = append(a.LLM.Message, openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				ToolCalls: toolCallParams,
			},
		})

		// Execute each tool call and append a tool-result message.
		for _, tc := range toolCalls {
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				errMsg := fmt.Sprintf("argument parse error: %v", err)
				fmt.Printf("[agent] step %d | tool %s | %s\n", step+1, tc.Function.Name, errMsg)
				a.LLM.Message = append(a.LLM.Message, openai.ToolMessage(errMsg, tc.ID))
				continue
			}

			toolResult, err := a.MCP.CallTool(tc.Function.Name, args)
			if err != nil {
				toolResult = fmt.Sprintf("tool call error: %v", err)
			}
			fmt.Printf("[agent] step %d | tool %s | result: %s\n", step+1, tc.Function.Name, toolResult)
			a.LLM.Message = append(a.LLM.Message, openai.ToolMessage(toolResult, tc.ID))
		}

		// Ask the model again with the tool results now in context.
		result, toolCalls = a.LLM.Chat("")
	}

	if len(toolCalls) > 0 {
		return result, fmt.Errorf("agent stopped after %d steps with unresolved tool calls", maxSteps)
	}
	return result, nil
}
