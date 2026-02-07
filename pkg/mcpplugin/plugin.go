package mcpplugin

import (
	"context"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type PluginTool struct {
	Tool    mcp.Tool
	Handler server.ToolHandlerFunc
}

type Plugin struct {
	Name    string
	Version string
	tools   map[string]*PluginTool
	lock    sync.Mutex
}

func (p *Plugin) AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	p.lock.Lock()
	p.tools[tool.Name] = &PluginTool{Tool: tool, Handler: handler}
	p.lock.Unlock()
}

// Notes:
// These implement a subset of the MCP Client functions for a plugin as we don;t need any of the fancy baclk and forth with
// an mcp server (http or otherwise)

// Initialize prepares the plugin with the provided context and initialization request.
func (p *Plugin) Initialize(ctx context.Context, request mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return &mcp.InitializeResult{}, nil
}

// ListTools retrieves the list of tools available in the plugin.
func (p *Plugin) ListTools(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	list := &mcp.ListToolsResult{
		Tools: make([]mcp.Tool, 0),
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	for _, tool := range p.tools {
		list.Tools = append(list.Tools, tool.Tool)
	}

	return list, nil
}

// CallTool calls a tool in the plugin.
func (p *Plugin) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tool := p.tools[request.Params.Name]
	return tool.Handler(ctx, request)
}

// Close cleans up any resources used by the plugin.
func (p *Plugin) Close() error {
	return nil
}

// New creates a new plugin instance.
func New(name string, version string) *Plugin {
	return &Plugin{
		Name:    name,
		Version: version,
		tools:   make(map[string]*PluginTool),
	}
}
