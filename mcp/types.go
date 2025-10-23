package mcp

import (
	"context"
	"encoding/json"
)

type ServerConfig struct {
	ProtocolVersion string
	Capabilities    Capabilities
	ServerInfo      ServerInfo
}

type Handler func(context context.Context, params json.RawMessage) (any, error)

type Server struct {
	Handlers     map[string]Handler
	serverConfig ServerConfig
	Tools        map[string]Tool
}

type initializedResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

type Capabilities struct {
	Tools *ToolsCapabilities `json:"tools,omitempty"`
}

type ToolsCapabilities struct {
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type toolsListResult struct {
	Tools []Tool `json:"tools"`
}

type Tool struct {
	Name        string                                                             `json:"name"`
	Description string                                                             `json:"description"`
	InputSchema InputSchema                                                        `json:"inputSchema"`
	Method      func(context context.Context, params json.RawMessage) (any, error) `json:"-"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type ToolCallResult struct {
	Content []ToolContent `json:"content"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type handleErrorOpts struct {
	code int
	msg  string
}
