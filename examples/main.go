package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	mcp "github.com/nadavushka-dev/mcpaws/mcp"
)

func executeGitStatus(ctx context.Context, params json.RawMessage) (any, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain=v1", "--branch")
	// cmd.Dir = "/Users/Nadavushka/code/personal/learn/mcp/git_mcp"

	output, err := cmd.Output()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return mcp.ToolCallResult{
				Content: []mcp.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Error: %s", string(exitErr.Stderr)),
				}},
			}, nil
		}
		return nil, fmt.Errorf("failed to execute git status: %w", err)
	}

	result := string(output)
	if result == "" {
		result = "Working directory is clean - no changes to commit."
	}

	return mcp.ToolCallResult{
		Content: []mcp.ToolContent{
			{
				Type: "text",
				Text: result,
			},
		},
	}, nil
}

func executeGitLog(ctx context.Context, args json.RawMessage) (any, error) {
	const (
		defaultLogCount = 3
		maxLogCount     = 50
	)

	var params struct {
		Count int `json:"count"`
	}

	params.Count = defaultLogCount

	if args != nil {
		err := json.Unmarshal(args, &params)
		if err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
	}

	if params.Count <= 0 {
		params.Count = defaultLogCount
	}
	if params.Count > maxLogCount {
		params.Count = maxLogCount
	}

	cmd := exec.CommandContext(ctx, "git", "log", "-n", fmt.Sprintf("%d", params.Count))

	output, err := cmd.Output()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return mcp.ToolCallResult{
				Content: []mcp.ToolContent{{
					Type: "text",
					Text: fmt.Sprintf("Error: %s", string(exitErr.Stderr)),
				}},
			}, nil
		}
		return nil, fmt.Errorf("failed to execute git log: %w", err)
	}

	result := string(output)
	if result == "" {
		result = "there are no logs"
	}

	return mcp.ToolCallResult{
		Content: []mcp.ToolContent{
			{
				Type: "text",
				Text: result,
			},
		},
	}, nil
}

func main() {
	s := mcp.NewMcpServer(mcp.ServerConfig{
		ProtocolVersion: "2025-06-18",
		Capabilities:    mcp.Capabilities{Tools: &mcp.ToolsCapabilities{}},
		ServerInfo: mcp.ServerInfo{
			Name:    "NEW-GIT-MCP",
			Version: "0.1.0",
		},
	})

	s.AddTool(mcp.Tool{
		Name:        "git_status",
		Description: "show current git status",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: make(map[string]mcp.Property),
			Required:   []string{},
		},
		Method: executeGitStatus,
	})

	s.AddTool(mcp.Tool{
		Name:        "git_log",
		Description: "Shows recent git commits with messages, authors, and dates",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"count": {
					Type:        "integer",
					Description: "number of logs to show (default is 3)",
					Default:     3,
				},
			},
			Required: []string{},
		},
		Method: executeGitLog,
	})

	s.Run()
}
