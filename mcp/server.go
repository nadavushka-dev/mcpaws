package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	jsonrpc "github.com/nadavushka-dev/mcpaws/jsonrpc"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if req.ID == nil {
		return jsonrpc.RPCResponse{}
	}
	handler := s.Handlers[req.Method]

	if handler == nil {
		return s.handleError(req.ID, nil, &handleErrorOpts{code: -32601, msg: "Method not found."})
	}
	result, err := handler(ctx, req.Params)

	if err != nil {
		return s.handleError(req.ID, err, nil)
	}

	resultByte, err := json.Marshal(result)
	if err != nil {
		return s.handleError(req.ID, err, &handleErrorOpts{code: -32700, msg: "Response marshaling failed."})
	}

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

func (s Server) handleError(id any, err error, opts *handleErrorOpts) jsonrpc.RPCResponse {
	config := handleErrorOpts{
		code: -32603,
		msg:  "",
	}

	if opts != nil {
		if opts.code != 0 {
			config.code = opts.code
		}
		if opts.msg != "" {
			config.msg = opts.msg
		}
	}

	message := config.msg
	if message == "" {
		if err != nil {
			message = err.Error()
		} else {
			message = "Something went terribly wrong"
		}
	}

	return jsonrpc.RPCResponse{
		Jsonrpc: "2.0",
		ID:      id,
		Error: &jsonrpc.RPCError{
			Code:    config.code,
			Message: message,
		},
	}
}
