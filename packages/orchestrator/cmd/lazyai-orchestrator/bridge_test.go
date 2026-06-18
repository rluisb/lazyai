package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type fakeUpstreamToolCaller struct {
	calls  []mcp.CallToolRequest
	result *mcp.CallToolResult
	err    error
}

func (f *fakeUpstreamToolCaller) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	f.calls = append(f.calls, request)
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func TestToolProxyServerListsAndCallsUpstreamTools(t *testing.T) {
	upstream := &fakeUpstreamToolCaller{result: mcp.NewToolResultText("ok")}
	proxy := newToolProxyServer(upstream, []mcp.Tool{{
		Name:        "list_catalog",
		Description: "List orchestration catalog definitions.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"kind": map[string]any{"type": "string"},
			},
		},
	}})

	listResponse := proxy.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
	listBytes, err := json.Marshal(listResponse)
	if err != nil {
		t.Fatalf("marshal tools/list response: %v", err)
	}
	var listed struct {
		Result mcp.ListToolsResult `json:"result"`
		Error  any                 `json:"error"`
	}
	if err := json.Unmarshal(listBytes, &listed); err != nil {
		t.Fatalf("unmarshal tools/list response: %v", err)
	}
	if listed.Error != nil {
		t.Fatalf("tools/list returned error: %v", listed.Error)
	}
	if len(listed.Result.Tools) != 1 || listed.Result.Tools[0].Name != "list_catalog" {
		t.Fatalf("unexpected tools/list result: %+v", listed.Result.Tools)
	}
	if listed.Result.Tools[0].Description != "List orchestration catalog definitions." {
		t.Fatalf("tool description was not preserved: %q", listed.Result.Tools[0].Description)
	}

	callResponse := proxy.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":"call-1","method":"tools/call","params":{"name":"list_catalog","arguments":{"kind":"chain"}}}`))
	callBytes, err := json.Marshal(callResponse)
	if err != nil {
		t.Fatalf("marshal tools/call response: %v", err)
	}
	var called struct {
		Result mcp.CallToolResult `json:"result"`
		Error  any                `json:"error"`
	}
	if err := json.Unmarshal(callBytes, &called); err != nil {
		t.Fatalf("unmarshal tools/call response: %v", err)
	}
	if called.Error != nil {
		t.Fatalf("tools/call returned error: %v", called.Error)
	}
	if len(upstream.calls) != 1 {
		t.Fatalf("expected one upstream call, got %d", len(upstream.calls))
	}
	if upstream.calls[0].Params.Name != "list_catalog" {
		t.Fatalf("tool name was not preserved: %q", upstream.calls[0].Params.Name)
	}
	args, ok := upstream.calls[0].Params.Arguments.(map[string]any)
	if !ok || args["kind"] != "chain" {
		t.Fatalf("tool arguments were not preserved: %#v", upstream.calls[0].Params.Arguments)
	}
	if len(called.Result.Content) != 1 {
		t.Fatalf("upstream result content was not returned: %+v", called.Result.Content)
	}
}

func TestToolProxyServerDoesNotFakeUnsupportedCapabilities(t *testing.T) {
	proxy := newToolProxyServer(&fakeUpstreamToolCaller{}, nil)

	response := proxy.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"resources/list","params":{}}`))
	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal resources/list response: %v", err)
	}
	var decoded struct {
		Error *mcp.JSONRPCErrorDetails `json:"error"`
	}
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("unmarshal resources/list response: %v", err)
	}
	if decoded.Error == nil {
		t.Fatalf("expected resources/list to return a JSON-RPC error")
	}
	if decoded.Error.Code != mcp.METHOD_NOT_FOUND {
		t.Fatalf("expected method not found, got %d (%s)", decoded.Error.Code, decoded.Error.Message)
	}
}

func TestToolProxyServerReturnsUpstreamCallErrors(t *testing.T) {
	proxy := newToolProxyServer(&fakeUpstreamToolCaller{err: errors.New("upstream exploded")}, []mcp.Tool{{
		Name:        "start_chain",
		Description: "Start a chain.",
	}})

	response := proxy.HandleMessage(context.Background(), json.RawMessage(`{"jsonrpc":"2.0","id":"call-1","method":"tools/call","params":{"name":"start_chain","arguments":{"chain":"demo","task":"ship"}}}`))
	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal tools/call response: %v", err)
	}
	var decoded struct {
		Result *mcp.CallToolResult      `json:"result"`
		Error  *mcp.JSONRPCErrorDetails `json:"error"`
	}
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("unmarshal tools/call response: %v", err)
	}
	if decoded.Error == nil {
		t.Fatalf("expected upstream error to be returned as JSON-RPC error, got success: %+v", decoded.Result)
	}
	if !strings.Contains(decoded.Error.Message, "upstream exploded") {
		t.Fatalf("upstream error message was not preserved: %+v", decoded.Error)
	}
	if decoded.Result != nil && len(decoded.Result.Content) > 0 {
		t.Fatalf("expected no successful tool content when upstream fails, got %+v", decoded.Result.Content)
	}
}

func TestReconnectNoopsWhenUpstreamWasAlreadyReplaced(t *testing.T) {
	current := &mcpclient.Client{}
	failed := &mcpclient.Client{}
	bridge := &stdioHTTPBridge{
		url:      "http://127.0.0.1:1/mcp",
		upstream: current,
	}

	if err := bridge.reconnect(context.Background(), failed); err != nil {
		t.Fatalf("reconnect should no-op when upstream changed: %v", err)
	}
	if bridge.currentUpstream() != current {
		t.Fatalf("reconnect replaced current upstream after another goroutine already did")
	}
}
