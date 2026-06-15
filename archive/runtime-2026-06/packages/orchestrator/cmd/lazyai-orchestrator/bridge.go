package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcptransport "github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const bridgeVersion = "0.1.0"

type upstreamToolCaller interface {
	CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

type stdioHTTPBridge struct {
	url string

	mu       sync.Mutex
	upstream *mcpclient.Client
	server   *server.MCPServer
}

func runStdioBridge(mcpURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	bridge, err := newStdioHTTPBridge(ctx, mcpURL)
	cancel()
	if err != nil {
		return err
	}
	defer bridge.Close()

	return server.ServeStdio(bridge.server)
}

func newStdioHTTPBridge(ctx context.Context, mcpURL string) (*stdioHTTPBridge, error) {
	bridge := &stdioHTTPBridge{url: mcpURL}

	upstream, err := connectUpstream(ctx, mcpURL)
	if err != nil {
		return nil, fmt.Errorf("initialize upstream MCP daemon: %w", err)
	}
	bridge.upstream = upstream

	toolsResult, err := upstream.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		_ = upstream.Close()
		return nil, fmt.Errorf("list upstream MCP tools: %w", err)
	}

	bridge.server = newToolProxyServer(bridge, toolsResult.Tools)
	return bridge, nil
}

func connectUpstream(ctx context.Context, mcpURL string) (*mcpclient.Client, error) {
	upstream, err := mcpclient.NewStreamableHttpClient(mcpURL)
	if err != nil {
		return nil, err
	}

	if err := upstream.Start(ctx); err != nil {
		_ = upstream.Close()
		return nil, err
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "lazyai-orchestrator-stdio-bridge",
		Version: bridgeVersion,
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	if _, err := upstream.Initialize(ctx, initRequest); err != nil {
		_ = upstream.Close()
		return nil, err
	}

	return upstream, nil
}

func newToolProxyServer(upstream upstreamToolCaller, tools []mcp.Tool) *server.MCPServer {
	proxy := server.NewMCPServer("lazyai-orchestrator", bridgeVersion,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	for _, upstreamTool := range tools {
		tool := upstreamTool
		if tool.Name == "" {
			continue
		}
		proxy.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			request.Request.Method = string(mcp.MethodToolsCall)
			return upstream.CallTool(ctx, request)
		})
	}

	return proxy
}

func (b *stdioHTTPBridge) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	failedUpstream := b.currentUpstream()
	result, err := callToolWithUpstream(ctx, failedUpstream, request)
	if err == nil {
		return result, nil
	}
	if !isUpstreamSessionLost(err) {
		return nil, err
	}

	if reconnectErr := b.reconnect(ctx, failedUpstream); reconnectErr != nil {
		return nil, fmt.Errorf("upstream MCP session was lost and reconnect failed: %w (original error: %v)", reconnectErr, err)
	}
	return b.callTool(ctx, request)
}

func (b *stdioHTTPBridge) callTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return callToolWithUpstream(ctx, b.currentUpstream(), request)
}

func (b *stdioHTTPBridge) currentUpstream() *mcpclient.Client {
	b.mu.Lock()
	defer b.mu.Unlock()
	upstream := b.upstream
	return upstream
}

func callToolWithUpstream(ctx context.Context, upstream *mcpclient.Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if upstream == nil {
		return nil, fmt.Errorf("upstream MCP client is not connected")
	}

	request.Request.Method = string(mcp.MethodToolsCall)
	return upstream.CallTool(ctx, request)
}

func (b *stdioHTTPBridge) reconnect(ctx context.Context, failedUpstream *mcpclient.Client) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.upstream != failedUpstream {
		return nil
	}
	if b.upstream != nil {
		_ = b.upstream.Close()
		b.upstream = nil
	}

	upstream, err := connectUpstream(ctx, b.url)
	if err != nil {
		return err
	}
	b.upstream = upstream
	return nil
}

func (b *stdioHTTPBridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.upstream == nil {
		return nil
	}
	err := b.upstream.Close()
	b.upstream = nil
	return err
}

func isUpstreamSessionLost(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, mcptransport.ErrSessionTerminated) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "Session not found") ||
		strings.Contains(msg, "session terminated") ||
		strings.Contains(msg, "HTTP 404") ||
		strings.Contains(msg, "404") ||
		strings.Contains(msg, "Not connected") ||
		strings.Contains(msg, "Transport is closed")
}
