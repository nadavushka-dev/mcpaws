package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *Server) handleInitialize(ctx context.Context, params json.RawMessage) (any, error) {
	return initializedResult(s.serverConfig), nil
}

func (s *Server) handleToolsList(ctx context.Context, params json.RawMessage) (any, error) {
	toolSlice := make([]Tool, 0, len(s.Tools))
	for key := range s.Tools {
		toolSlice = append(toolSlice, s.Tools[key])
	}

	return &toolsListResult{Tools: toolSlice}, nil
}

func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (any, error) {
	var callParams toolCallParams
	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	tool, exists := s.Tools[callParams.Name]

	if !exists {
		return nil, fmt.Errorf("Unknown tool %s", callParams.Name)
	}

	return tool.Method(ctx, callParams.Arguments)
}
