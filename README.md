# Go MCP Wrapper

A lightweight Go library for building Model Context Protocol (MCP) servers.

## Installation

```bash
go get github.com/nadavushka-dev/go_mcp_wrapper
```

## Usage

```go
package main

import (
    "encoding/json"
    "github.com/nadavushka-dev/go_mcp_wrapper/mcp"
)

func main() {
    server := mcp.NewMcpServer(mcp.ServerConfig{
        ProtocolVersion: "2025-06-18",
        Capabilities:    mcp.Capabilities{Tools: &mcp.ToolsCapabilities{}},
        ServerInfo: mcp.ServerInfo{
            Name:    "my-server",
            Version: "1.0.0",
        },
    })
    
    server.AddTool(mcp.Tool{
        Name:        "echo",
        Description: "Echo back a message",
        InputSchema: mcp.InputSchema{
            Type: "object",
            Properties: map[string]mcp.Property{
                "message": {Type: "string", Description: "Message to echo"},
            },
            Required: []string{"message"},
        },
        Method: func(params json.RawMessage) (any, error) {
            var args struct{ Message string `json:"message"` }
            json.Unmarshal(params, &args)
            return mcp.ToolCallResult{
                Content: []mcp.ToolContent{{Type: "text", Text: args.Message}},
            }, nil
        },
    })
    
    server.Run()
}
```

## Example

See [`examples/main.go`](examples/main.go) for a Git integration example.

## License

MIT License