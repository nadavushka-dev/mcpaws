package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestHandleInitialize(t *testing.T) {
	s := &Server{
		serverConfig: ServerConfig{
			ProtocolVersion: "2025-06-18",
			Capabilities:    Capabilities{Tools: &ToolsCapabilities{}},
			ServerInfo:      ServerInfo{Name: "TestServer", Version: "1.0.0"},
		},
	}

	ctx := context.Background()
	result, err := s.handleInitialize(ctx, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	initResult, ok := result.(initializedResult)
	if !ok {
		t.Fatalf("Expected initializedResult, got %T", result)
	}

	if initResult.ProtocolVersion != "2025-06-18" {
		t.Errorf("Expected protocol version '2025-06-18', got %s", initResult.ProtocolVersion)
	}
}

func TestHandleToolsList(t *testing.T) {
	s := &Server{
		Tools: map[string]Tool{"test_tool": {Name: "test_tool", Description: "Tool for testings"}},
	}

	ctx := context.Background()
	result, err := s.handleToolsList(ctx, nil)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	listResult, ok := result.(*toolsListResult)
	if !ok {
		t.Fatalf("Expected *toolsListResult, got %T", result)
	}

	if len(listResult.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(listResult.Tools))
	}
}

func TestHandleToolsCall(t *testing.T) {
	tool := Tool{
		Name:        "Mock_Tool",
		Description: "Some mock tool for testings",
		Method: func(ctx context.Context, params json.RawMessage) (any, error) {
			// Expect only the arguments ({}), not the full params
			if string(params) != "{}" {
				return nil, fmt.Errorf("expected arguments only, got: %s", string(params))
			}
			return ToolCallResult{Content: []ToolContent{{Type: "text", Text: "mock result"}}}, nil
		}}

	ctx := context.Background()

	s := &Server{
		Tools: map[string]Tool{"Mock_Tool": tool},
	}

	params := `{"name": "Mock_Tool", "arguments": {}}`

	result, err := s.handleToolsCall(ctx, json.RawMessage(params))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	callResult, ok := result.(ToolCallResult)
	if !ok {
		t.Fatalf("Expected ToolCallResult, got %T", result)
	}

	if callResult.Content[0].Text != "mock result" {
		t.Errorf("Expected 'mock result', got %s", callResult.Content[0].Text)
	}
}
