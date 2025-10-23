package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	jsonrpc "github.com/nadavushka-dev/mcpaws/jsonrpc"
)

func TestNewMcpServer(t *testing.T) {
	conf := ServerConfig{
		ProtocolVersion: "2025-06-18",
		Capabilities:    Capabilities{Tools: &ToolsCapabilities{}},
		ServerInfo:      ServerInfo{Name: "TestServer", Version: "1.0.0"},
	}

	s := NewMcpServer(conf)

	if s.serverConfig.ProtocolVersion != "2025-06-18" {
		t.Errorf("Protocol version does not match, wanted 2025-06-18, got %s", s.serverConfig.ProtocolVersion)
	}
	if s.Handlers == nil || s.Tools == nil {
		t.Error("Handles or Tools map not initialized")
	}
	if _, exists := s.Handlers["initialize"]; !exists {
		t.Error("initialize handler not set")
	}
}

func TestAddHandler(t *testing.T) {
	s := NewMcpServer(ServerConfig{
		ProtocolVersion: "2025-06-18",
		Capabilities:    Capabilities{Tools: &ToolsCapabilities{}},
		ServerInfo:      ServerInfo{Name: "TestServer", Version: "1.0.0"},
	})

	s.AddHandler("CustomHandler", func(context context.Context, params json.RawMessage) (any, error) { return nil, nil })

	_, exists := s.Handlers["CustomHandler"]
	if !exists {
		t.Error("Handler did not added to server")
	}
}

func TestProcessRequest(t *testing.T) {
	s := &Server{
		Handlers: map[string]Handler{
			"test_method": func(ctx context.Context, params json.RawMessage) (any, error) {
				return "success", nil
			},
		},
		Tools: make(map[string]Tool),
	}

	tests := []struct {
		name     string
		req      jsonrpc.RPCRequest
		expected jsonrpc.RPCResponse
		hasError bool
	}{
		{
			name: "valid request",
			req: jsonrpc.RPCRequest{
				Jsonrpc: "2.0",
				ID:      "test-id",
				Method:  "test_method",
			},
			expected: jsonrpc.RPCResponse{
				Jsonrpc: "2.0",
				ID:      "test-id",
				Result:  json.RawMessage(`"success"`),
			},
			hasError: false,
		},
		{
			name: "nil ID",
			req: jsonrpc.RPCRequest{
				Jsonrpc: "2.0",
				ID:      nil,
				Method:  "test_method",
			},
			expected: jsonrpc.RPCResponse{},
			hasError: false,
		},
		{
			name: "unknown method",
			req: jsonrpc.RPCRequest{
				Jsonrpc: "2.0",
				ID:      "test-id",
				Method:  "unknown",
			},
			expected: jsonrpc.RPCResponse{
				Jsonrpc: "2.0",
				ID:      "test-id",
				Error:   &jsonrpc.RPCError{Code: -32601, Message: "Method not found."},
			},
			hasError: false,
		},
		{
			name: "handler error",
			req: jsonrpc.RPCRequest{
				Jsonrpc: "2.0",
				ID:      "test-id",
				Method:  "error_method",
			},
			expected: jsonrpc.RPCResponse{
				Jsonrpc: "2.0",
				ID:      "test-id",
				Error:   &jsonrpc.RPCError{Code: -32603, Message: "handler failed"},
			},
			hasError: true,
		},
	}

	// Add error handler for the error test case
	s.Handlers["error_method"] = func(ctx context.Context, params json.RawMessage) (any, error) {
		return nil, errors.New("handler failed")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.processRequest(tt.req)

			if tt.expected.Jsonrpc == "" && result.Jsonrpc != "" {
				t.Errorf("Expected empty response, got %v", result)
			} else if tt.expected.Jsonrpc != "" {
				if result.Jsonrpc != tt.expected.Jsonrpc {
					t.Errorf("Expected jsonrpc %s, got %s", tt.expected.Jsonrpc, result.Jsonrpc)
				}
				if result.ID != tt.expected.ID {
					t.Errorf("Expected ID %v, got %v", tt.expected.ID, result.ID)
				}
				if tt.expected.Error != nil {
					if result.Error == nil {
						t.Errorf("Expected error, got nil")
					} else if result.Error.Code != tt.expected.Error.Code {
						t.Errorf("Expected error code %d, got %d", tt.expected.Error.Code, result.Error.Code)
					} else if result.Error.Message != tt.expected.Error.Message {
						t.Errorf("Expected error message '%s', got '%s'", tt.expected.Error.Message, result.Error.Message)
					}
				}
				if tt.expected.Result != nil && string(result.Result) != string(tt.expected.Result) {
					t.Errorf("Expected result %s, got %s", string(tt.expected.Result), string(result.Result))
				}
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	s := Server{}

	tests := []struct {
		name     string
		id       any
		err      error
		opts     *handleErrorOpts
		expected jsonrpc.RPCResponse
	}{
		{
			name: "with error",
			id:   "test-id",
			err:  errors.New("test error"),
			opts: nil,
			expected: jsonrpc.RPCResponse{
				Jsonrpc: "2.0",
				ID:      "test-id",
				Error:   &jsonrpc.RPCError{Code: -32603, Message: "test error"},
			},
		},
		{
			name: "with custom opts",
			id:   "test-id",
			err:  nil,
			opts: &handleErrorOpts{code: -32601, msg: "Custom message"},
			expected: jsonrpc.RPCResponse{
				Jsonrpc: "2.0",
				ID:      "test-id",
				Error:   &jsonrpc.RPCError{Code: -32601, Message: "Custom message"},
			},
		},
		{
			name: "nil error and no message",
			id:   "test-id",
			err:  nil,
			opts: nil,
			expected: jsonrpc.RPCResponse{
				Jsonrpc: "2.0",
				ID:      "test-id",
				Error:   &jsonrpc.RPCError{Code: -32603, Message: "Something went terribly wrong"},
			},
		},
		{
			name: "custom code only",
			id:   "test-id",
			err:  errors.New("original error"),
			opts: &handleErrorOpts{code: -32700, msg: ""},
			expected: jsonrpc.RPCResponse{
				Jsonrpc: "2.0",
				ID:      "test-id",
				Error:   &jsonrpc.RPCError{Code: -32700, Message: "original error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.handleError(tt.id, tt.err, tt.opts)

			if result.Jsonrpc != tt.expected.Jsonrpc {
				t.Errorf("Expected jsonrpc %s, got %s", tt.expected.Jsonrpc, result.Jsonrpc)
			}
			if result.ID != tt.expected.ID {
				t.Errorf("Expected ID %v, got %v", tt.expected.ID, result.ID)
			}
			if tt.expected.Error != nil {
				if result.Error == nil {
					t.Errorf("Expected error, got nil")
				} else {
					if result.Error.Code != tt.expected.Error.Code {
						t.Errorf("Expected error code %d, got %d", tt.expected.Error.Code, result.Error.Code)
					}
					if result.Error.Message != tt.expected.Error.Message {
						t.Errorf("Expected error message '%s', got '%s'", tt.expected.Error.Message, result.Error.Message)
					}
				}
			}
		})
	}
}
