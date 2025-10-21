package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	jsonrpc "github.com/nadavushka-dev/go_mcp_wrapper/jsonrpc"
)

func NewMcpServer(conf ServerConfig) *Server {
	s := &Server{
		Handlers: make(map[string]Handler),
	}

	s.serverConfig = conf

	s.Handlers["initialize"] = s.handleInitialize
	s.Handlers["call"] = s.handleToolsCall
	s.Handlers["tools/list"] = s.handleToolsList

	s.Tools = make(map[string]Tool)

	return s
}

func (s *Server) AddHandler(name string, handler Handler) {
	s.Handlers[name] = handler
}

func (s *Server) AddTool(tool Tool) {
	s.Tools[tool.Name] = tool
}

func (s Server) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		line := scanner.Text()

		if line == "" {
			continue
		}

		var req jsonrpc.RPCRequest
		err := json.Unmarshal([]byte(line), &req)

		if err != nil {
			s.sendResponse(jsonrpc.RPCResponse{
				Jsonrpc: "2.0",
				ID:      nil,
				Error: &jsonrpc.RPCError{
					Code:    -32700,
					Message: "Parse error",
				},
			})
			continue
		}

		response := s.processRequest(req)
		if response.Jsonrpc != "" {
			s.sendResponse(response)
		}
	}
}

func (s Server) processRequest(req jsonrpc.RPCRequest) jsonrpc.RPCResponse {
	if req.ID == nil {
		return jsonrpc.RPCResponse{}
	}
	handler := s.Handlers[req.Method]

	if handler == nil {
		return jsonrpc.RPCResponse{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Error: &jsonrpc.RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
	result, err := handler(req.Params)

	if err != nil {
		return jsonrpc.RPCResponse{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Error: &jsonrpc.RPCError{
				Code:    -32603,
				Message: err.Error(),
			},
		}
	}

	resultByte, _ := json.Marshal(result)

	return jsonrpc.RPCResponse{
		Jsonrpc: "2.0",
		ID:      req.ID,
		Result:  resultByte,
	}

}

func (s Server) sendResponse(res jsonrpc.RPCResponse) {
	resBytes, err := json.Marshal(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal response %v\n", err)
		return
	}

	fmt.Println(string(resBytes))
}
