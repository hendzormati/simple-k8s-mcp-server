package tools

import (
    "github.com/mark3labs/mcp-go/mcp"
)

// ListPodsTool creates a tool for listing pods in a namespace
func ListPodsTool() mcp.Tool {
    return mcp.NewTool(
        "listPods",
        mcp.WithDescription("List all pods in a Kubernetes namespace"),
        mcp.WithString("namespace", mcp.Description("The namespace to list pods from (default: 'default')")),
    )
}